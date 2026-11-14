# Epic 2: Decks & Spaced-Repetition Study (Anki-style engine)

| Field | Value |
|---|---|
| Status | Implemented (2026-11-14) on branch `feat/epic2-srs-study-engine`. Scheduler: go-fsrs v4 (FSRS-6). Open questions resolved with the proposals: forward cards only, image on the back (front optional), FSRS options under "Advanced", no gamification yet. Notes: "Study all" applies each deck's own limits (like Anki), with no extra global cap; example translations are shown directly under each example. |
| Depends on | Epic 1 |
| Blocks | Epic 3 (cards need decks), Epic 4 |
| Research | [docs/features/vocabulary/anki-srs-research.md](../../docs/features/vocabulary/anki-srs-research.md) |

## 1. Goal

Users can organise vocabulary cards into decks and review them every day with a scheduler that behaves like Anki: the same states, buttons, daily limits, learning steps and FSRS intervals. The study screen is richer than Anki, with audio and an image on each card.

This epic delivers the **engine and the study experience**. Automatic card content comes in Epic 3. Here, cards can be created with a simple manual form so the engine can be built and tested on its own.

## 2. Scope

**In scope**
- Deck management (create, rename, delete, per-deck daily limits).
- Card storage with scheduling state, plus a manual "add card" form (basic fields).
- FSRS scheduler as a pure domain service, with learning and relearning steps, fuzz, leeches and a day boundary.
- Study queue, answer, undo, and preview of the next interval for each button.
- Study UI: front/back flip, audio, image, keyboard shortcuts, session summary.
- Card browser: list, search, edit, suspend, reset.
- Basic stats: today's counts, streak, 7-day due forecast.

**Out of scope (later)**
- Per-user FSRS parameter optimisation (we collect review logs now so we can add it later).
- `.apkg` import/export to Anki.
- Cloze cards, typed-answer cards.
- Offline study (PWA) and syncing across devices while offline.

## 3. Domain model

```
Deck 1───* Card *───1 Lexeme (optional, from Epic 3)
             │
             └───* ReviewLog
```

- **Card content is a snapshot.** When a card is created from a shared `lexeme` (Epic 3/4), its content JSON is copied into the card. Users can then edit their own card without changing anyone else's, and we keep `lexeme_id` for de-duplication and stats.
- **Card content schema**: defined in Epic 3 §4. In this epic only these fields are required: `headword`, `meaning_vi`; optional: `ipa`, `example_en`, `image_url`, `audio_url`.

### Tables

```sql
decks(
  id uuid pk, user_id fk, name text, description text,
  source enum('manual','paste','collection') default 'manual',
  collection_id uuid null,                 -- Epic 4
  new_per_day int null, reviews_per_day int null,   -- null = use user default
  archived_at, created_at, updated_at)

cards(
  id uuid pk, user_id fk, deck_id fk, lexeme_id uuid null,
  content jsonb not null, content_version int,
  -- scheduling state (FSRS)
  state enum('new','learning','review','relearning') default 'new',
  due timestamptz not null,               -- new cards: created_at (queue order uses position)
  position int,                           -- order for new cards inside the deck
  stability double precision, difficulty double precision,
  step smallint default 0,                -- current (re)learning step index
  reps int default 0, lapses int default 0,
  last_review_at timestamptz,
  scheduled_days int, elapsed_days int,
  suspended bool default false, buried_until date null, is_leech bool default false,
  created_at, updated_at,
  unique(user_id, deck_id, lexeme_id))    -- no duplicate lexeme in a deck

review_logs(
  id uuid pk, card_id fk, user_id fk,
  rating smallint check (rating between 1 and 4),
  state_before, state_after,
  due_before, due_after,
  stability_before, stability_after, difficulty_before, difficulty_after,
  step_before, elapsed_days, scheduled_days,
  duration_ms int, reviewed_at timestamptz,
  undone_at timestamptz null)

-- indexes: cards(user_id, deck_id, state, due) where not suspended; review_logs(user_id, reviewed_at)
```

## 4. Scheduler (backend `study` module)

- `IScheduler` interface in `app/modules/study/domain/`. It is pure, does no I/O, and receives a clock and a random source.
  ```go
  type IScheduler interface {
      Preview(card entity.Card, now time.Time) map[Rating]ScheduledResult // for button labels
      Answer(card entity.Card, rating Rating, now time.Time) (entity.Card, entity.ReviewLog)
  }
  ```
- Implementation wraps `go-fsrs` (FSRS-6, default parameters). Configuration comes from user settings and deck overrides: `desired_retention`, `learning_steps`, `relearning_steps`, `max_interval`, `enable_fuzz`.
- **Day boundary:** "today" is computed in the user's timezone with `day_start_hour` (default 4:00). Review cards are due when `due < start of tomorrow`.
- **Leech:** when `lapses` reaches 8, set `is_leech=true`. The UI shows a hint to improve the card (change the image or add a mnemonic). Leech cards are not auto-suspended.
- **Fuzz** is disabled in tests through config, so scheduling scenarios are deterministic.

## 5. Study queue algorithm

For a request `GET /study/queue?deck_id=<id|all>`:

1. Compute `today_start` and `tomorrow_start` for the user.
2. Count what was already introduced or reviewed today (from `review_logs`) and subtract it from the limits.
3. Build the queue in this order:
   1. Learning/relearning cards with `due <= now + 20m` (learn-ahead), ordered by due.
   2. Review cards with `due < tomorrow_start`, ordered by due, limited to the remaining review limit.
   3. New cards ordered by `position`, limited to the remaining new-card limit, interleaved with reviews (1 new after every N reviews, where N = reviews/new).
4. Return the first batch (≤ 20 cards) with the full content, the next-interval preview for each button, and the totals `{new, learning, review}`.
5. The client asks for more when it runs low. Learning cards answered during the session come back when they are due. The server is the source of truth; the client never computes intervals.

**Edge cases**
- Studying across the day boundary at 4:00 refreshes the limits.
- Two tabs open: answering is idempotent on `(card_id, client_review_id)`. A stale answer, where the card state changed since it was fetched, returns `409 STUDY_CARD_STATE_CHANGED` and the client refetches.
- Nothing due: show "Congratulations, you're done for today". Offer "Study more new cards (+10)" (raises today's limit) and "Review ahead".

## 6. API

| Method | Path | Purpose |
|---|---|---|
| GET | `/decks` | List decks with counts `{new, learning, review_due, total}` |
| POST | `/decks` | Create a deck `{name, description?}` |
| PATCH | `/decks/:id` | Rename or change limits |
| DELETE | `/decks/:id` | Soft-delete (archive) a deck and its cards |
| GET | `/decks/:id/cards` | Card browser: `?q=&state=&leech=&page=` |
| POST | `/decks/:id/cards` | Manual card `{content}` |
| PATCH | `/cards/:id` | Edit content or move to another deck |
| POST | `/cards/:id/suspend` · `/unsuspend` · `/reset` | Card actions (reset = back to new) |
| GET | `/study/queue?deck_id=` | Queue batch + counts (see §5) |
| POST | `/study/answer` | `{card_id, rating, duration_ms, client_review_id}` returns the updated card and new counts |
| POST | `/study/undo` | Undo the last answer (restored from the review log) |
| GET | `/stats/overview` | Today's reviews and new cards, streak, forecast for the next 7 days, retention over the last 30 days |

The study routes are rate-limited (medium-term limiter).

## 7. User stories

### US-2.1 Manage decks
- [ ] Create, rename and delete decks. Deletion asks for confirmation and shows the number of cards.
- [ ] The deck list shows `new · learning · due` counts using Anki colours mapped to our tokens (`--srs-new`, `--srs-learn`, `--srs-review`).
- [ ] A "Study" button per deck, plus a global **"Study all due"** button.
- [ ] Optional per-deck override of new cards/day and reviews/day.

### US-2.2 Add a card manually
- [ ] A form with headword, Vietnamese meaning, optional IPA, optional example, optional image URL.
- [ ] A duplicate headword in the same deck shows a warning.
- [ ] *(Epic 3 replaces this as the main entry point; the form stays as "Add custom card".)*

### US-2.3 Study session
**As a** learner **I want** to review today's cards one by one **so that** I remember words for the long term.

- [ ] **Front**: headword (large), IPA, a speaker button, and the part of speech. Optional setting: show the image on the front, for picture-first recognition.
- [ ] **Reveal** with Space, Enter or a tap on "Show answer". The back slides in below the front, as in Anki, so the question stays visible.
- [ ] **Back**: image, Vietnamese meaning, word forms (each with one example), collocations, and extra examples. Every example sentence has a 🔊 button.
- [ ] **Audio**: if `audio_autoplay` is on, the word plays on show and again on reveal. If the card has no server audio yet (before Epic 3), use the browser **Web Speech API** (`en-GB` or `en-US` voice) as a fallback.
- [ ] **Buttons** Again / Hard / Good / Easy show the next-interval preview (`<1m`, `<6m`, `<10m`, `3d`), with shortcuts 1–4. On mobile they are large thumb-reachable buttons.
- [ ] Header shows remaining counts `new · learning · review` and a progress bar.
- [ ] **Undo** last answer (Ctrl/⌘+Z or a button).
- [ ] Per-card menu: edit card, suspend, "mark as leech fixed", report wrong content.
- [ ] Measure time on card (`duration_ms`), capped at 60 s, as Anki does.
- [ ] **Session summary**: cards studied, time spent, accuracy (Good+Easy share), and tomorrow's due count.
- [ ] Images are lazy-loaded with fixed aspect-ratio placeholders, so the buttons never move.

### US-2.4 Card browser
- [ ] Search by headword or meaning; filter by state (new / learning / review / suspended / leech).
- [ ] Columns: headword, meaning, state, due (relative), interval, lapses.
- [ ] Bulk actions: suspend, unsuspend, move to deck, reset, delete.

### US-2.5 Study settings
- [ ] User-level settings (from Epic 1) are editable: new/day, reviews/day, auto-play audio, accent, show image on front.
- [ ] An "Advanced" section: desired retention (slider 80–97% with an explanation: higher means more reviews), learning steps, relearning steps, maximum interval.
- [ ] Changing settings never rewrites existing due dates; they apply from the next answer onwards.

### US-2.6 Dashboard & stats (basic)
- [ ] Dashboard hero: "**42 cards due today**" with a "Start studying" button and today's progress.
- [ ] Streak: consecutive days with at least one review, using the user's day boundary.
- [ ] A 7-day due forecast bar chart and the true retention rate over the last 30 days.

## 8. Test plan

- **Scheduler unit tests** (table-driven, fuzz off, fixed clock):
  - new → Good → (10m) → Good → graduates with interval ≈ initial stability (FSRS default for Good ≈ 2–3d).
  - new → Again → stays at step 0 (1m); Hard on step 0 → about 6m; Easy → graduates immediately with the longest interval.
  - review → Again → relearning, lapses+1, stability drops; then Good → back to review with a short interval.
  - Interval ordering for a review card: `Again < Hard < Good < Easy`.
  - Late review (due 10 days ago) → longer interval than on-time (because R is lower).
  - Leech at the 8th lapse.
- **Queue tests**: daily limits respected after the 4:00 boundary across timezones (UTC+7, UTC−5); learn-ahead window; suspended and buried cards excluded.
- **API tests**: idempotent answers, 409 on stale state, undo restores the exact previous state.
- **E2E**: study 5 cards with the keyboard only; counts update; summary shown.

## 9. Open questions for review

1. **Reverse cards** (Vietnamese → English, i.e. "produce the word"): add as an optional second card per word? This is better for IELTS Writing and Speaking (active vocabulary) but doubles the review load. *Proposal: v1 is forward only; a deck setting "Also create reverse cards" comes later.*
2. **Image placement**: back only (as specified), with an optional setting to show it on the front? *Proposal: yes, default off.*
3. **Show desired retention to normal users**, or hide all FSRS settings behind "Advanced"? *Proposal: hide.*
4. **Gamification** (XP, streak freeze, leaderboards): out of scope for now?

## 10. Definition of done

- A user can create a deck, add cards manually, and study them over several simulated days (test clock), with intervals matching the FSRS reference implementation within fuzz tolerance.
- All UI strings are in vi and en; the study screen works one-handed on a 375px-wide phone.
- The scheduler package has ≥ 90% unit-test coverage.
