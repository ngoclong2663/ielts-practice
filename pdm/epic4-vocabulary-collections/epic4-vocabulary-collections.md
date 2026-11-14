# Epic 4: Vocabulary Collections ("Learn N words today")

| Field | Value |
|---|---|
| Status | Draft, awaiting review |
| Depends on | Epic 2 (decks/study), Epic 3 (lexemes, generation pipeline) |

## 1. Goal

Offer curated word lists such as an **IELTS 4000-word collection**. A learner picks a collection, enters a **title** and **how many words to learn today**, and immediately gets a deck of rich cards (same format as Epic 3), ready to study with the Anki-style engine (Epic 2). Each day they come back for the next batch while reviewing earlier words.

## 2. Scope

**In scope**
- Collection catalogue (browse, details, progress).
- "Start today's words" dialog: title, word count, order.
- Creating a daily deck from the next unseen words.
- "I already know this word" marking.
- Per-collection progress.
- An admin **seed pipeline** that imports word lists and pre-generates all lexemes (content, image, audio) before publishing.
- Topic sub-groups inside a collection (e.g. Environment, Education, Technology).

**Out of scope**
- User-created public collections and sharing.
- Admin web UI (v1 uses CLI scripts plus a `needs_review` query).
- Placement test to skip known words automatically (later).

## 3. User flow

```
[Collections] ─► [Collection detail: description, topics, progress 120/4000]
                        │
                        └─ "Learn new words" ─► Dialog:
                               Title:          [IELTS 4000 · Day 7 · 14/11]
                               Words today:    [ 20 ]  (5–50, slider + input)
                               Order:          (•) In list order  ( ) By topic [Environment ▼]  ( ) Random
                               ℹ 20 new words/day ≈ 150–200 reviews/day after a few weeks
                               [Preview words]  [Create & start]
                        │
                        ▼
              [Preview list: 20 words, ✓ "I know this" toggles; known words are replaced by the next ones]
                        │
                        ▼
              New deck created → Study session starts (Epic 2)
```

## 4. Behaviour rules

- **Next words** = items of the collection in the chosen order, excluding lexemes that the user:
  - already has as a card in **any** deck,
  - marked as **known** in this collection,
  - or skipped (also counts as known).
- **Each "Learn new words" action creates a new deck** with the user's title. The deck stores `source='collection'` and `collection_id`, and the deck list groups these decks under their collection.
- The new deck's `new_per_day` is set to the chosen count, so **all chosen words are introduced today** regardless of the user's global new-card limit. Reviews of older decks keep their own schedule.
- **Default title:** `{collection short name} · Day {n} · {dd/MM}`, localised. Day `n` = the number of decks already created from this collection + 1. The title is editable and required (1–80 characters).
- **Default count:** the last count used for this collection, otherwise 20.
- The user may create several batches on the same day. Show a soft warning when today's total new words exceeds 50.
- **Review load hint:** estimate steady-state reviews ≈ 8–10 × new words per day and show it in the dialog so users choose sustainably.
- If a lexeme in the batch is not `ready` (not pre-generated yet), create the deck anyway: the card appears with a "Preparing…" state and the Epic 3 job pipeline fills it. This should be rare, since collections are pre-generated before publishing.
- "Study all due in this collection" studies across all decks of the collection.

## 5. Data model

```sql
collections(
  id uuid pk, slug text unique,
  title_vi text, title_en text, short_name text,
  description_vi text, description_en text,
  level text,                    -- e.g. 'IELTS 5.0–6.5'
  cover_image_url text, word_count int,
  status enum('draft','published','archived'),
  license_note text,             -- source list + license
  created_at, updated_at)

collection_items(
  collection_id fk, lexeme_id fk, position int,
  topic text null, band text null,  -- e.g. 'core', 'academic', 'advanced'
  primary key (collection_id, lexeme_id))
-- index (collection_id, position), (collection_id, topic, position)

user_collection_known(
  user_id, collection_id, lexeme_id, marked_at,
  primary key (user_id, collection_id, lexeme_id))

-- decks.collection_id (added in Epic 2) links daily decks to the collection
```

**Progress**, computed per user and collection (cached and refreshed on answer):

| Metric | Definition |
|---|---|
| Added | Distinct collection lexemes the user has as cards |
| Learning | Those cards in learning, relearning, or review with interval < 21d |
| Mature | Review cards with interval ≥ 21d (Anki's definition) |
| Known | Marked "I know this" |
| Remaining | word_count − added − known |

## 6. API

| Method | Path | Purpose |
|---|---|---|
| GET | `/collections` | Published collections with the user's progress summary |
| GET | `/collections/:slug` | Details, topics with counts, progress, last used count |
| GET | `/collections/:slug/next-words?count=&order=&topic=` | Preview the next N words (headword, meaning, image thumbnail) |
| POST | `/collections/:slug/known` | `{lexeme_ids[]}` mark as known; `DELETE` to undo |
| POST | `/collections/:slug/batches` | `{title, count, order, topic?, exclude_lexeme_ids?}` → `{deck_id}` (transactional) |
| GET | `/collections/:slug/decks` | The user's decks created from this collection |

`POST /batches` is a single transaction: select the next words, create the deck, then create the cards (content snapshots from lexemes). It rejects with `COLLECTION_NO_WORDS_LEFT` when the collection is exhausted.

## 7. User stories

### US-4.1 Browse collections
- [ ] A grid of collection cards: cover, localised title, level, word count, and the user's progress ring ("120 / 4000").
- [ ] The collection detail page shows the description, topic chips with counts, a progress breakdown, and the decks created so far.

### US-4.2 Start today's words
- [ ] Clicking **"Learn new words"** opens a dialog (shadcn `Dialog`, RHF + Zod) asking for the **title** and the **number of words** (plus order).
- [ ] Validation: title 1–80 characters; count 5–50 (configurable) and ≤ remaining words.
- [ ] The review-load hint updates live as the count changes.
- [ ] **Create & start** makes the deck and opens the study session within 2 s (pre-generated content).

### US-4.3 Preview & mark known words
- [ ] **Preview words** lists the next N words with meaning and thumbnail.
- [ ] Toggling "I know this" replaces the word with the next candidate, keeping N constant.
- [ ] Known words never appear again in this collection; they can be un-marked on the collection page.

### US-4.4 Progress & continuity
- [ ] The dashboard shows "Continue IELTS 4000: Day 8" as a quick action once today's reviews are done.
- [ ] Collection progress updates after study sessions.
- [ ] Decks from a collection are grouped in "My decks", and one button studies all due cards across them.

### US-4.5 Seed pipeline (admin, CLI)
- [ ] `make seed-collection FILE=collections/ielts-4000.csv`. The CSV has columns `position, headword, sense_hint, topic, band`.
- [ ] Creates or updates the collection (as a `draft`) and its items, upserting lexemes by key.
- [ ] Generates missing lexeme content through the **Message Batches API** (asynchronous, 50% cheaper than real-time calls; well suited to thousands of words), then images and TTS through the same Epic 3 steps, with rate limiting for image providers.
- [ ] Writes a QA report: failures, invalid words, missing images, and words flagged `visualizable=false`.
- [ ] `make publish-collection SLUG=...` publishes only when 100% of lexemes are `ready`, or the operator passes `--allow-partial`.
- [ ] Re-running the seed is idempotent.

## 8. Content & licensing ⚠️

- *"4000 Essential English Words"* is a **copyrighted** book series. We must **not** copy its list, ordering, examples or images.
- **Proposal:** build our own IELTS collection from openly licensed or freely usable lists, then curate and order it for IELTS:
  - **NGSL** (New General Service List, ~2,800 core words) and **NAWL** (New Academic Word List, ~960 academic words), published under CC BY-SA. Verify the current licence terms before use.
  - **AWL** (Academic Word List, Coxhead, 570 word families). Check its usage terms.
  - Topic vocabulary for common IELTS themes, written by our team.
- All example sentences, meanings and images are generated or licensed by us (Epic 3), so the card content is our own.
- Store the source and licence for each collection in `license_note` and show it on the collection page if the licence requires attribution.

## 9. Proposed initial collections

| Collection | Size | Notes |
|---|---|---|
| IELTS Core 4000 | ~4,000 | Base NGSL + NAWL, ordered by frequency, tagged by topic |
| IELTS Academic Word List | ~570 families | For Reading and Writing Task 2 |
| IELTS Topic Vocabulary | ~1,000 | 20 topics × ~50 words (Environment, Education, Health, Technology, ...) |

The first release can ship **one** collection (IELTS Core) to limit generation cost and QA effort.

## 10. Open questions for review

1. **Collection source and name:** do you have a specific "4000 IELTS words" list you own or have rights to? If not, are you OK with building it from NGSL/NAWL plus curation (proposal above)?
2. **One deck per day** (your description, proposed) vs a single deck per collection where each "learn today" just adds N new cards? Per-day decks give a clear sense of daily achievement; a single deck keeps "My decks" tidier. *Proposal: per-day decks, grouped visually under the collection.*
3. **Word count bounds:** 5–50 per batch? Any daily maximum?
4. **Order options:** are list order, by topic and random all needed in v1, or only list order?
5. **"I know this":** should known words count toward the progress ring as "learned"?

## 11. Definition of done

- One published collection whose words are all pre-generated with content, image and audio.
- A user can open it, enter a title and a count of 20, preview, mark 2 words as known (replaced automatically), create the deck and study it immediately.
- The next day, "Learn new words" continues from where they stopped, without repeats, and progress is correct.
- All UI in vi and en.
