# Research: How Anki Schedules Vocabulary Reviews

> Purpose: understand Anki's repetition logic well enough to rebuild it faithfully for our vocabulary feature (Epic 2), then extend it with audio and images (Epic 3).
> Date: 2026-11-14

---

## 1. Core concepts

| Anki concept | Meaning | Our equivalent |
|---|---|---|
| **Note** | One piece of content (e.g. the word "decide" with all its fields) | `lexeme` (global, shared content) |
| **Card** | One reviewable question generated from a note (front → back). A note can produce several cards (forward, reverse) | `card` (per user, holds the scheduling state) |
| **Deck** | A container of cards, with its own daily limits and options | `deck` |
| **Preset (deck options)** | Settings shared by decks: limits, learning steps, scheduler params | User defaults + per-deck overrides |
| **Review log** | One row per answer; used for stats, undo and FSRS optimisation | `review_log` |

AnkiWeb (the browser version) runs the same scheduler as Anki desktop, because both use the same Rust core. The logic below applies to both.

### Card states

```
            ┌──────────── Again ─────────────┐
            ▼                                │
 NEW ──► LEARNING ──(passes last step)──► REVIEW ──(Again = lapse)──► RELEARNING
                                            ▲                              │
                                            └────(passes relearn steps)────┘
```

- **New**: the user has never seen the card.
- **Learning**: the card is being shown again within minutes or hours, following the *learning steps*.
- **Review**: the card has "graduated" and is shown again after days, weeks or months.
- **Relearning**: a review card the user forgot (pressed *Again*). It goes through *relearning steps* and then returns to Review.
- Extra flags: **suspended** (never shown), **buried** (hidden until tomorrow), **leech** (forgotten too often).

### The four answer buttons

| Button | Key | Meaning |
|---|---|---|
| Again | 1 | Forgot / wrong |
| Hard | 2 | Recalled with serious difficulty |
| Good | 3 | Recalled correctly |
| Easy | 4 | Trivially easy |

Anki shows the next interval above each button (for example `<1m · <6m · <10m · 4d`). We will do the same.

---

## 2. A study day

1. **Day boundary.** A new "Anki day" starts at a configurable hour (default **4:00 AM local time**), so studying past midnight still counts as the same day.
2. **Daily limits.** Each deck has **New cards/day** (default 20) and **Maximum reviews/day** (default 200). Cards in learning are not limited.
3. **Building the queue.** Anki gathers, in priority order:
   1. **Intraday learning cards** that are due now. "Learn ahead" also shows learning cards due within the next 20 minutes if nothing else is left.
   2. **Interday learning cards** (learning steps of 1 day or more), mixed with reviews.
   3. **Review cards** due today, oldest first, up to the review limit.
   4. **New cards**, up to the new-card limit. By default they are mixed in with reviews.
4. **Counts.** The header shows three numbers: `new · learning · review` (blue · red · green).
5. **Siblings.** Cards from the same note can be buried until the next day so you don't see "decide → quyết định" and "quyết định → decide" in the same session. This matters only if we add reverse cards.

---

## 3. Learning phase (both schedulers use this)

Default **learning steps**: `1m 10m`. Default **relearning steps**: `10m`.

| Answer while in learning | Result |
|---|---|
| Again | Go back to the first step (show again in 1m) |
| Hard | Repeat the current step. On the first step, use the average of step 1 and step 2 (about 6m) |
| Good | Move to the next step. After the last step, the card **graduates** to Review |
| Easy | Graduate immediately with a longer first interval |

- **Graduating interval** (SM-2): 1 day. **Easy interval**: 4 days.
- With FSRS, the first interval after graduating comes from the card's initial *stability* instead (see §5).

---

## 4. Review phase: legacy SM-2 (Anki's default until FSRS)

Each card has an **ease** factor, which starts at 250%, and an **interval** in days.

| Answer | New interval | Ease change |
|---|---|---|
| Again | Card lapses and goes to relearning. After relearning: `max(1d, interval × newIntervalPct)` (default 0%, so 1 day) | −20% |
| Hard | `interval × 1.2` | −15% |
| Good | `(interval + daysLate/2) × ease` | none |
| Easy | `(interval + daysLate) × ease × 1.3` (easy bonus) | +15% |

- Every interval is also multiplied by the *interval modifier* (default 1.0). It is clamped to at least the previous interval + 1 day, and to at most the *maximum interval* (default 36,500 days).
- Ease never drops below 130%.
- **Weakness, known as "ease hell":** repeated *Again* and *Hard* answers keep lowering ease, so a card can get stuck at short intervals forever. This is one reason FSRS replaced SM-2.

---

## 5. Review phase: FSRS (Anki's modern scheduler)

FSRS (Free Spaced Repetition Scheduler) was added to Anki in 23.10 and is now the recommended scheduler (FSRS-6 in current releases). It models memory with three variables (the "DSR" model):

| Variable | Meaning |
|---|---|
| **D: Difficulty** (1–10) | How hard this card is for this user |
| **S: Stability** (days) | How many days until the chance of recalling the card falls to 90% |
| **R: Retrievability** (0–1) | The chance of recalling the card *right now* |

**Forgetting curve** (power law, FSRS-6):

```
R(t, S) = (1 + factor · t / S) ^ (−w20),   where factor = 0.9^(−1/w20) − 1   (so R(S, S) = 0.9)
```

**Next interval** for a target *desired retention* `r` (default **0.90**):

```
I(r, S) = S / factor · ( r^(−1/w20) − 1 )        → with r = 0.9, I ≈ S
```

**After each answer, FSRS updates D and S:**

- **Initial state** (first answer on a new card): `S0 = w[G−1]`. With the FSRS-6 default parameters this is roughly Again ≈ 0.2d, Hard ≈ 1.3d, Good ≈ 2.3d, Easy ≈ 8.3d. Initial difficulty is `D0(G) = w4 − e^(w5·(G−1)) + 1`.
- **Difficulty:** `D' = D + (−w6·(G−3)) · (10−D)/9`, then pulled back toward `D0(4)` by `w7` (mean reversion). Again and Hard raise difficulty; Easy lowers it.
- **Stability after a successful recall** (Hard, Good or Easy) grows more when:
  - the card is easier (low D),
  - the current S is small (S^−w9: short memories grow faster), and
  - R was low at review time (reviewing just before forgetting gives the biggest boost; this is the "desirable difficulty" effect).

  Hard applies a penalty (`w15`) and Easy applies a bonus (`w16`):
  `S' = S · (e^w8 · (11−D) · S^(−w9) · (e^(w10·(1−R)) − 1) · hardPenalty · easyBonus + 1)`
- **Stability after a lapse** (Again) drops to a post-lapse stability:
  `S'f = min(S, w11 · D^(−w12) · ((S+1)^w13 − 1) · e^(w14·(1−R)))`
- **Same-day reviews** (learning steps) also adjust S slightly (FSRS-6: `w17`–`w19`).
- **21 parameters** (`w0..w20`) have good defaults. Anki can **optimise** them for each user from their review logs (recommended after about 1,000 reviews).

**Why FSRS over SM-2:**

- **Fewer reviews.** FSRS's published benchmarks show it reaches the same retention with fewer reviews than SM-2.
- **One intuitive setting.** The user picks a *desired retention* (for example 90%) instead of tuning ease and modifiers.
- **No ease hell.**
- **Late or early reviews are handled correctly**, because R is computed from the actual elapsed time.

---

## 6. Other behaviours worth copying

| Behaviour | Anki default | Notes |
|---|---|---|
| **Fuzz** | On | Adds a small random spread (about ±5%) to intervals of 3 days or more, so cards learned together don't stay clumped together forever |
| **Leech** | 8 lapses → tag "leech" (option: auto-suspend) | Surfaces words that need a better card (a better image or mnemonic) |
| **Undo** | Last answer | Needs the previous card state, which we store in the review log |
| **Mature card** | Interval ≥ 21 days | Used in stats: "learned" vs "mature" |
| **Review ≈ 10× new** | Rule of thumb | 20 new/day leads to roughly 150–200 reviews/day in steady state. We should warn users who pick a large daily word count |

---

## 7. Decisions for our product

| Topic | Decision (proposed) |
|---|---|
| Scheduler | **FSRS-6** with default parameters. SM-2 is not implemented |
| Library | `github.com/open-spaced-repetition/go-fsrs/v4` v4.0.0 (official Go port, FSRS-6, learning steps, fuzz, rollback). Wrapped in `be/app/modules/study/domain/scheduler.go` so it can be swapped |
| Desired retention | 0.90 default, adjustable 0.80–0.97 under "Advanced" |
| Learning steps | `1m 10m`; relearning `10m` |
| Daily limits | 20 new / 200 reviews (user default). Collections override "new" with the count the user picks that day |
| Day boundary | 4:00 AM in the user's timezone |
| Max interval | 36,500 days (Anki default), configurable |
| Fuzz | On in production, off in tests (deterministic) |
| Leech | Threshold 8, tag only (no auto-suspend); suggest editing the card |
| Review logs | Store every answer from day 1, so we can add per-user FSRS optimisation later |
| Card types | v1: one forward card per word (English → meaning). Reverse card is an open question (Epic 2) |

### What we add on top of Anki

- **Audio:** word pronunciation (UK/US) and example sentences read aloud, with optional auto-play on reveal.
- **Image:** an illustration for each word (Unsplash or Pixabay), with photographer attribution.
- **Rich back side:** word forms, Vietnamese meaning, collocations, and one example per word form, generated automatically (Epic 3).
- **Curated collections:** "learn N words today" from IELTS word lists (Epic 4).

---

## Sources

- Anki Manual, Deck Options: https://docs.ankiweb.net/deck-options.html
- FSRS for Anki (open-spaced-repetition): https://github.com/open-spaced-repetition/fsrs4anki
- FSRS algorithm wiki: https://github.com/open-spaced-repetition/fsrs4anki/wiki/The-Algorithm
- go-fsrs: https://github.com/open-spaced-repetition/go-fsrs
- Anki release notes: https://github.com/ankitects/anki/releases
- Unsplash API docs and guidelines: https://unsplash.com/documentation
- Pixabay API docs: https://pixabay.com/api/docs/
