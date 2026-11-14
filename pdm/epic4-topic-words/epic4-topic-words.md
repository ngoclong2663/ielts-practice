# Epic 4: Words by Topic ("Give me 15 words about hospitals")

| Field | Value |
|---|---|
| Status | Implemented (2026-11-14) on branch `feat/epic4-topic-words`. Replaces the earlier "Vocabulary Collections" plan (curated lists), which is dropped. |
| Depends on | Epic 2 (decks, study), Epic 3 (generation pipeline, lexeme cache, review grid) |

## 1. Goal

On **Add words**, a learner picks a **topic** (e.g. Hospital, Money, Environment) and **how many words**, and gets a list of **new** words for that topic: a mix of everyday, IELTS and TOEIC words. The list leaves out words already in the chosen deck and words the learner has marked as known. The learner reviews the list, keeps the words they want, and those words then go through the existing Epic 3 flow (generate cards → review cards → save to the deck).

Why this replaces collections:
- No curated "4000 IELTS words" list to source or license.
- It reuses the Epic 3 pipeline end to end. Only the "which words?" step is new.
- Learners study the topics they need (IELTS Writing Task 2 and Speaking are organised by topic).

## 2. Scope

**In scope**
- A third tab on Add words: **By topic** (next to "Paste a list" and "Add one by hand").
- Preset IELTS topics plus a free-text topic.
- Word count (5–30). No level choice: every list mixes everyday, IELTS and TOEIC words.
- Word suggestions from the LLM, with duplicates and known words removed on the server.
- A word-list review step: keep/remove each word, "I know this", "Suggest more".
- Hand-off to the Epic 3 job for the kept words, targeting the chosen deck.
- Removing the Collections feature (menu item, page, placeholder).

**Out of scope**
- Curated or public word lists, sharing, admin tools.
- Automatic level detection (placement test).
- Daily "learn N words from a list" scheduling. The deck's own new-cards-per-day limit (Epic 2) already paces learning.

## 3. User flow

```
[Add words] ─ tabs: Paste a list | By topic | Add one by hand
                        │
                        ▼ By topic
   Topic:   [Health] [Money] [Environment] [Education] [Technology] … [Other: ________]
   Words:   [ 15 ]  (5–30)
   Save to: [My hospital words ▼]  or create a new deck
   [Suggest words]
                        │
                        ▼
   Word list (15 new words; 4 already in this deck were skipped)
     ☑ diagnosis     n.   chẩn đoán            B2   [I know this]
     ☑ admit         v.   nhập viện, tiếp nhận  B2   [I know this]
     ☐ nurse         n.   y tá                  A2   (unchecked by the user)
     …
   [Suggest more]                                   [Generate 14 cards]
                        │
                        ▼
   Epic 3 review grid (progress → preview/edit/picture) → Save to deck → Study now
```

## 4. Behaviour rules

- **Topics.** The user either **picks a preset** (one click) or **types their own topic** in the "Other" field (decided: both). Presets cover the common IELTS themes:
  - Health & hospital, Money & finance, Environment, Education, Technology, Work & careers, Crime & law, Travel & tourism, Food, Media & advertising, Housing & cities, Family & society, Sport, Science.
  - Each preset carries an English description that is sent to the model, e.g. *"Health & hospital: hospitals, illnesses, treatment, public health"*.
  - Free text is limited to 60 characters and treated as data in the prompt, never as instructions.
- **No level choice** (changed in review). Each list is a varied mix, roughly one third each:
  - everyday words for the topic;
  - IELTS words, for Speaking and Writing Task 2;
  - TOEIC words, for work, business and customer service.
  - Difficulty is mixed (mostly CEFR A2–C1), without trivially basic words. Each suggested word shows its own CEFR badge.
- **Word mix.** Mostly single words useful for IELTS Writing and Speaking, plus a few common collocations or phrasal verbs (at most about 20%). No proper nouns, no very rare or technical words.
- **Duplicates.** The server removes a suggestion when its lemma (or its lexeme key, `headword|sense|vi`) matches:
  - a card in the **chosen deck** (by `lexeme_id` or a case-insensitive headword match). Other decks are not checked (decided);
  - a word the user marked **known** (§5);
  - a word suggested in the same session that the user removed (so "Suggest more" never repeats it).
  - The UI shows how many were skipped ("4 already in this deck").
- **Exactly N.** The model is asked for about 1.5×N candidates plus a short exclusion list (§6.2). The server filters them and returns the first N. If fewer than N remain, the list says so and offers "Suggest more".
- **Review step.**
  - Every suggestion starts checked; the user unchecks words they don't want.
  - "I know this" also unchecks the word and stores it as known.
  - "Suggest more" adds N more words to the same list.
  - The Generate button shows the count of checked words, at most 30, the Epic 3 job limit.
- **Hand-off.** "Generate N cards" creates a normal Epic 3 job with `entries = checked words` and `target_deck_id = chosen deck`. Everything after that is Epic 3: the cache, quota, pictures, review, and saving with duplicate skipping.
- **Cost and quota.**
  - A suggestion request is one small LLM call: a list of words with a short meaning, no full cards.
  - It does **not** count toward the daily generated-word quota. It is limited separately to **5 suggestion requests per account per day** (decided). A new list and each "Suggest more" both count; the UI shows how many are left. The limit is a setting (`VOCAB_DAILY_SUGGESTIONS`, default 5).
  - Card generation afterwards uses the normal Epic 3 quota, and words already in the shared cache are still free.

## 5. Data model

```sql
-- Words a user said they already know: never suggested again.
user_known_words(
  user_id uuid references users on delete cascade,
  lemma text,                      -- lowercase dictionary form
  created_at timestamptz,
  primary key (user_id, lemma))

-- One suggestion request: rate limit, "suggest more" continuity, and debugging prompt quality.
topic_suggestions(
  id uuid pk, user_id uuid, deck_id uuid null,
  topic text, requested int,           -- (a `level` column was dropped in migration 000006)
  words jsonb,                     -- [{lemma, pos, meaning_vi, cefr, kept}]
  skipped_existing int,
  input_tokens int, output_tokens int,
  created_at timestamptz)
```

No changes to cards, decks or lexemes. The unused `decks.collection_id` column and the `'collection'` deck source stay for now; dropping them gains nothing.

## 6. Backend (`vocab` module, same LLM providers as Epic 3)

### 6.1 API

| Method | Path | Purpose |
|---|---|---|
| GET | `/vocab/topics` | Preset topics (id, en/vi label, description) |
| POST | `/vocab/suggestions` | `{topic_id \| topic_text, count, deck_id?}` → `{id, words[], skipped_existing}` |
| POST | `/vocab/suggestions/:id/more` | `{count, removed[]}` → more words, never repeating earlier ones |
| POST | `/vocab/known-words` | `{lemma}`: mark as known (DELETE to undo) |

The hand-off uses the existing `POST /vocab/jobs` with `target_deck_id`.

### 6.2 Suggestion prompt

- Same providers as Epic 3: Claude (`VOCAB_LLM_MODEL`) first, Gemini fallback. Structured JSON output: `{words: [{lemma, pos, meaning_vi, cefr}]}`.
- Input: the topic description, the candidate count, and an **exclusion list**. The list holds the lemmas in the chosen deck plus known words, capped at the 300 most recent so the prompt stays small. The server filters again anyway, so the cap only reduces wasted suggestions.
- The prompt asks for topic-specific words, mixing everyday, IELTS and TOEIC vocabulary and difficulty levels, none on the exclusion list.

### 6.3 Filtering

1. Normalise each suggestion with the Epic 3 rules (lowercase, length and word-count limits).
2. Drop exclusions (deck cards, known words, earlier suggestions in the session) and duplicates within the answer.
3. Keep the first N. Store the request in `topic_suggestions`.

## 7. Web

- **Add words tabs:** Paste a list · **By topic** · Add one by hand.
- **By topic form:**
  - topic chips with an "Other" text field;
  - a count stepper (5–30);
  - the deck picker shared with "Paste a list": the last used deck, or create one.
- **Word list:**
  - each row has a checkbox, the word, part of speech, a short Vietnamese meaning and a CEFR badge, plus a speaker button (browser speech) and "I know this";
  - a sticky footer holds "Suggest more" and "Generate N cards".
- **Leaving the page:** the open suggestion list is kept in the URL (`?suggestion=<id>`), so a refresh does not lose it.
- **Collections removed:** the Collections item leaves the side and bottom navigation (the bottom nav goes back to 4 items), the `/collections` page is deleted, and the `proxy.ts` path entry is removed.
- All new strings in vi and en.

## 8. User stories

### US-4.1 Suggest words for a topic
- [ ] Pick a preset or type a topic, choose a count (5–30), choose the deck.
- [ ] Get exactly N new words in under 10 s (p90), or fewer with a message when the topic runs out.
- [ ] The form shows how many suggestion requests are left today (5 per day); at 0 the button is disabled with a message.
- [ ] The UI shows how many words were skipped because they are already in the deck.

### US-4.2 Review the word list
- [ ] Uncheck words; "I know this" hides the word now and in future suggestions (undo available in the toast).
- [ ] "Suggest more" adds new words and never repeats one already shown in this list.
- [ ] Each word has a short Vietnamese meaning and can be read aloud.

### US-4.3 Generate cards from the list
- [ ] "Generate N cards" opens the Epic 3 review grid for exactly the checked words, targeting the chosen deck.
- [ ] Saving skips any word added to the deck in the meantime (existing Epic 3 behaviour).

### US-4.4 Remove Collections
- [ ] No Collections link, page or route remains; the mobile bottom navigation has 4 items.

## 9. Decisions (review 2026-11-14)

1. **Duplicates:** checked against the **chosen deck only**.
2. **"I know this":** remembered per user; those words are never suggested again.
3. **Level:** removed after review. Lists mix everyday, IELTS and TOEIC words of varied difficulty.
4. **Topics:** preset chips **or** a typed topic. Presets as listed in §4 (Health & hospital stays one preset; "hospital" can also be typed).
5. **Suggestion limit:** **5 requests per account per day**, counting new lists and "Suggest more".

## 10. Definition of done

- Choosing "Health & hospital" and 15 words for a deck that already holds 4 of the likely words returns 15 new, topic-relevant words (a mix of everyday, IELTS and TOEIC) and reports the 4 skipped words.
- Keeping 12 of them produces 12 cards through the Epic 3 review grid, saved into that deck.
- "I know this" words never come back for that user.
- The Collections feature is gone from navigation and routes.
- All new UI strings are in vi and en.
