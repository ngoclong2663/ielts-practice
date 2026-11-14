# Epic 3: Smart Card Generation (paste words → rich cards with audio & images)

| Field | Value |
|---|---|
| Status | Draft, awaiting review |
| Depends on | Epic 1, Epic 2 (decks and cards) |
| Blocks | Epic 4 (collections reuse this pipeline) |

## 1. Goal

A user pastes one or more English words and gets complete, ready-to-study cards in seconds. The **back** of each card has:

- word forms (each with **one example sentence**),
- the Vietnamese meaning,
- collocations,
- an **image** that illustrates the word (Unsplash or Pixabay),
- **audio**: the word and its example sentences read aloud.

Traditional Anki cards have no audio or images unless users build them by hand. Removing that work is the core value of the product.

## 2. Scope

**In scope**
- "Add words" screen: paste, normalise, preview.
- An asynchronous generation pipeline: LLM content, then image search, then text-to-speech.
- A shared **lexeme cache**, so a word is generated once and reused for every user.
- Review and edit before saving: change the image, edit text, regenerate one word.
- Saving the cards into an existing or new deck (they become **new** cards in Epic 2).
- Attribution and licence compliance for images.
- Cost controls: quotas, caching, logging token usage.

**Out of scope**
- AI-generated illustrations (possible later for abstract words).
- Generating from a photo or PDF (OCR).
- Meaning languages other than Vietnamese (the schema supports it; the UI doesn't yet).

## 3. User flow

```
[Add words] ──paste──► [Normalise & preview chips] ──Generate──► [Progress per word]
                                                                     │
          ┌──────────────────────────────────────────────────────────┘
          ▼
[Review grid: card preview per word] ── edit / swap image / regenerate / remove
          │
          └── Save to deck (existing or new) ──► Deck page / "Start studying now"
```

1. **Paste**: a textarea accepts words separated by new lines, commas or semicolons. Multi-word phrases stay together ("take into account", "carbon footprint"). Optional sense hint: `bank | river` or `bank (river)`.
2. **Normalise** (client and server): trim, collapse spaces, lowercase unless it's a proper noun, remove duplicates, and drop entries longer than 60 characters or with more than 6 words, since those look like sentences. Show chips with a ✕ to remove. **Limit: 30 entries per request.**
3. **Generate**: create a job. Words already in the lexeme cache are ready instantly; the rest are processed in parallel.
4. **Progress**: each chip moves through `queued → writing content → finding image → recording audio → ready`, or `failed` with a retry button.
5. **Review**: a grid of mini card previews. Each one has:
   - "Did you mean *accommodate*?" when the LLM flags a typo or non-word,
   - lemma suggestion: "You entered *running*; create a card for *run*?",
   - an image picker (6 candidates plus "search again" with a custom keyword),
   - inline edit of any field,
   - "Regenerate" for that word only.
6. **Save**: choose the target deck (defaults to the last used deck, or "New deck: Vocabulary 14/11/2026"). Cards are created in a single transaction. Words already in that deck are skipped and reported.

## 4. Card content schema (`content` JSON, version 1)

```jsonc
{
  "headword": "decide",
  "lemma": "decide",
  "sense_hint": null,
  "ipa": { "uk": "/dɪˈsaɪd/", "us": "/dɪˈsaɪd/" },
  "part_of_speech": "verb",
  "cefr": "B1",                                  // optional, for difficulty badge
  "meaning_vi": "quyết định",
  "definition_en": "to choose something after thinking about it",   // short, learner-friendly
  "word_forms": [                                 // one example per word form
    { "form": "decide",    "pos": "verb",      "ipa": "/dɪˈsaɪd/",   "meaning_vi": "quyết định",
      "example": { "en": "She decided to study abroad.", "vi": "Cô ấy quyết định đi du học." } },
    { "form": "decision",  "pos": "noun",      "ipa": "/dɪˈsɪʒn/",   "meaning_vi": "sự quyết định",
      "example": { "en": "It was a difficult decision.", "vi": "Đó là một quyết định khó khăn." } },
    { "form": "decisive",  "pos": "adjective", "ipa": "/dɪˈsaɪsɪv/", "meaning_vi": "quyết đoán",
      "example": { "en": "We need decisive action on climate change.", "vi": "..." } },
    { "form": "decisively","pos": "adverb",    "ipa": "...",         "meaning_vi": "một cách quyết đoán",
      "example": { "en": "...", "vi": "..." } }
  ],
  "collocations": [                               // 3–6, IELTS-useful
    { "phrase": "make a decision", "meaning_vi": "đưa ra quyết định" },
    { "phrase": "decide against something", "meaning_vi": "quyết định không làm gì" }
  ],
  "extra_examples": [                             // 1–2 extra example sentences, IELTS-style (academic or topic-based)
    { "en": "Governments must decide how to allocate limited resources.", "vi": "..." }
  ],
  "synonyms": ["determine", "resolve"],           // optional (see open questions)
  "other_senses": [ { "pos": "verb", "meaning_vi": "phân xử (tòa án)" } ],  // short list, optional
  "image": {
    "provider": "unsplash|pixabay|none", "provider_id": "abc",
    "url": "...", "thumb_url": "...", "width": 1080, "height": 720,
    "alt": "a person standing at a fork in the road",
    "author_name": "Jane Doe", "author_url": "...", "source_url": "..."
  },
  "audio": {
    "headword": { "uk": "media/audio/decide-uk.mp3", "us": "media/audio/decide-us.mp3" },
    "examples": { "wf0": "media/audio/<hash>.mp3", "x0": "..." }
  },
  "meta": { "meaning_lang": "vi", "generator": "claude-opus-5-5", "content_version": 1 }
}
```

**Interpretation of the requirement "various examples, one example per word":** every **word form** gets exactly one example, and the headword also gets 1–2 extra IELTS-style examples. See open question 1.

## 5. Generation pipeline (backend `vocab` module)

### 5.1 Lexeme cache (shared across users)

```sql
lexemes(
  id uuid pk,
  key text unique,             -- normalized: lower(headword) || '|' || coalesce(sense_hint,'') || '|' || meaning_lang
  headword text, lemma text, sense_hint text, meaning_lang text default 'vi',
  content jsonb, content_version int,
  status enum('pending','ready','failed','needs_review'),
  quality_flags text[], report_count int default 0,
  created_at, updated_at)
```

- If a lexeme with the key exists and is `ready`, it is reused (no LLM, image or TTS calls).
- Users edit **their card copy** (Epic 2 snapshot), never the shared lexeme.
- "Report wrong content" increments `report_count`. Past a threshold the lexeme becomes `needs_review`.

### 5.2 Jobs

```sql
generation_jobs(id, user_id, target_deck_id null, status enum('queued','running','done','partial','failed'),
                total int, done int, created_at, finished_at)
generation_job_items(id, job_id, input_text, sense_hint, lexeme_id null,
                     status enum('queued','content','image','audio','ready','failed','invalid'),
                     error_code, suggestion text, attempts int, updated_at)
```

- **Worker:** a goroutine pool inside the API process (size configurable, default 4). It claims items with `SELECT … FOR UPDATE SKIP LOCKED` and retries with exponential backoff, up to 3 attempts. This needs no Redis in v1; it can move to a separate worker binary later without changing the schema.
- **Progress:** the frontend polls `GET /vocab/jobs/:id` every 1.5 s with React Query `refetchInterval`, and stops when the status is final. SSE is an optional later improvement.

### 5.3 Step A: content (LLM)

- **Provider:** Anthropic Claude via the official Go SDK (`anthropic-sdk-go`).
  - Model: `claude-opus-5-5`, with `output_config.effort: "low"` (short, well-defined extraction task).
  - The model name is a config value (`VOCAB_LLM_MODEL`), so a cheaper model can be swapped in later if quality holds on our eval set.
- **Structured outputs** (`output_config.format` with a JSON Schema that matches §4, minus image and audio) guarantee parseable output. A validator also checks counts (word forms ≥ 1, collocations 3–6) and that example sentences actually contain the form.
- **Batching:** one request handles up to 10 words, returning an array keyed by input. This lowers per-word overhead.
- **Prompt caching:** the long, stable system prompt (instructions, style guide, two worked examples) goes first with `cache_control`. Only the word list varies.
- **Refusal handling:** check `stop_reason` and enable server-side `fallbacks: "default"` (beta `server-side-fallback-2026-07-01`).
- **The LLM also returns:**
  - `is_valid` and `suggestion` (typo or non-word detection),
  - `lemma`,
  - `image_query`: a concrete, visual search phrase. Abstract words search badly, e.g. "nevertheless" or "decide" → "person at a crossroads choosing a path".
  - `visualizable` (bool).
- **Prompt-injection safety:** pasted text is wrapped as data in the prompt. Entries are length-limited and the output is schema-constrained, so injected instructions cannot change the output shape.
- Every call is logged with token usage and cost per job, to track cost per word.

### 5.4 Step B: image

`IImageProvider` interface with two adapters. Provider order is configurable.

| | **Unsplash** | **Pixabay** |
|---|---|---|
| Rate limit | 50 req/h (demo) → 5,000 req/h after production approval | 100 req / 60 s |
| Storage rule | **Must hotlink** Unsplash image URLs (do not re-host) | **No permanent hotlinking**: download the chosen image to our storage. API responses must be cached for 24 h |
| Attribution | Required: "Photo by *Name* on Unsplash", with UTM-tagged links | Show Pixabay as the source (credit appreciated) |
| Extra obligation | Call the photo's `download_location` endpoint when the user selects the image for a card | Follow the Pixabay Content License |
| Quality | Better artistic photos | More illustrations and vectors (good for abstract words) |

- **Proposal:** Pixabay is primary during development (higher limit), with `image_type=all` and `safesearch=true`. Switch Unsplash to primary after applying for production access.
- Flow: search `image_query` → take the top 6 → auto-select #1 → store the candidates in the job item so the picker opens instantly.
- If `visualizable=false` or there are no results: use a neutral topic illustration or no image (the card layout must handle both).
- Pixabay images are downloaded, resized to WebP (max 800px wide) and stored at `media/images/<lexeme_id>.webp`.
- The attribution line is rendered under the image on the back of the card.

### 5.5 Step C: audio (TTS)

- Generate MP3 for the **headword (UK + US)** and **every example sentence** once per lexeme. Store at `media/audio/<sha1(text|voice)>.mp3`, so identical sentences share a file. Serve through a CDN with long cache headers.
- `ITtsProvider` interface. Candidate providers: Google Cloud TTS, Azure Speech or AWS Polly (neural voices, en-GB and en-US). See open question 3.
- The browser Web Speech API stays as a **fallback** whenever an audio file is missing.

## 6. API

| Method | Path | Purpose |
|---|---|---|
| POST | `/vocab/normalize` | `{text}` → `{entries:[{text, sense_hint}], rejected:[...]}` (instant, no LLM) |
| POST | `/vocab/jobs` | `{entries:[{text, sense_hint?}], target_deck_id?}` → `{job_id}` |
| GET | `/vocab/jobs/:id` | Job status + items (with content preview once ready) |
| POST | `/vocab/jobs/:id/items/:item_id/retry` | Retry or regenerate one item, optionally `{sense_hint}` |
| GET | `/vocab/jobs/:id/items/:item_id/images?q=` | More image candidates or a custom keyword |
| PATCH | `/vocab/jobs/:id/items/:item_id` | Pending edits before save: `{content_patch, image}` |
| POST | `/vocab/jobs/:id/commit` | `{deck_id | new_deck_name, item_ids[]}` → `{deck_id, created, skipped_duplicates}` |
| POST | `/lexemes/:id/report` | `{reason}` |

Rate limits: `POST /vocab/jobs` is limited per user per minute, and a **daily quota of uncached words** applies per user (see open question 4). Cache hits never count toward the quota.

## 7. User stories

### US-3.1 Paste words
- [ ] Textarea with placeholder examples; paste or type; live chip preview after normalisation; counter "12 / 30".
- [ ] Invalid entries (too long, sentence-like) are shown greyed out with a reason.
- [ ] The **Generate** button is disabled while there are 0 valid entries.

### US-3.2 Generate with live progress
- [ ] Cached words are ready in under 1 s.
- [ ] Uncached: p90 under 15 s for a single word; under 60 s for 20 words.
- [ ] A failed item shows a translated reason and a "Retry" button. Other items are not blocked.
- [ ] Leaving the page doesn't lose the job: "Recent generations" lists unfinished jobs.

### US-3.3 Review & edit
- [ ] The card preview is identical to the back of the card in the study screen (shared component).
- [ ] Typo or lemma suggestions can be accepted with one click, which regenerates that item.
- [ ] Image picker: 6 thumbnails, a custom search box, a "No image" option, and the attribution shown.
- [ ] Any text field can be edited inline; audio is regenerated for edited example sentences when saving.
- [ ] Remove an item from the batch.

### US-3.4 Save to deck
- [ ] Choose an existing deck or create one inline.
- [ ] Duplicate lexemes already in the deck are skipped, with a toast such as "3 words already in this deck".
- [ ] After saving: "Start studying now" opens the Epic 2 study session for that deck.

### US-3.5 Content quality & reporting
- [ ] "Report wrong content" on any card, with reasons (wrong meaning, bad example, wrong image, wrong audio).
- [ ] Maintain an **eval set** of about 100 words (common, academic, phrasal verbs, polysemous, abstract). A script runs the generator and checks the schema, Vietnamese meaning accuracy (manual spot check) and example validity. It runs whenever the prompt or model changes.

## 8. Non-functional requirements

- Content generation is idempotent per lexeme key: concurrent requests for the same new word create one lexeme (unique key plus `ON CONFLICT`).
- External API keys live only on the backend; the frontend never calls the LLM, Unsplash, Pixabay or TTS directly.
- Media is served from object storage through a CDN; image `srcset` uses the thumbnail and the full size.
- Observability: per-step latency, provider error rate, cache-hit rate, and LLM cost per day.

## 9. Open questions for review

1. **"One example per word"**: is my interpretation right (one example for each **word form**, plus 1–2 extra examples for the headword)? Or did you mean exactly one example per pasted word?
2. **Polysemous words** (*bank*, *address*, *issue*): one card for the most IELTS-relevant sense plus a short `other_senses` list (proposed), or one card per sense?
3. **TTS provider**: Google Cloud TTS, Azure Speech or AWS Polly? Or start with the browser Web Speech API only, which is free but voice quality varies by device? *Proposal: a cloud neural TTS with per-lexeme caching. The cost is one-time per word, not per user.*
4. **Quotas**: how many **uncached** words can a free user generate per day? *Proposal: 50/day, cache hits unlimited.*
5. **Synonyms/antonyms**: include them? They are useful for IELTS paraphrasing but make the card longer. *Proposal: include up to 3 synonyms, collapsed by default.*
6. **Vietnamese translation of example sentences**: show it always, hidden behind a tap (better for learning), or not at all? *Proposal: hidden behind a tap.*
7. **Image provider priority**: Pixabay first (proposed) or Unsplash first?

## 10. Definition of done

- Pasting 20 mixed words (including one typo, one phrase and one abstract word) produces 20 reviewable cards with meaning, word forms, collocations, examples, an image (or a deliberate "no image") and playable audio, saved into a deck and studyable in Epic 2.
- Image attribution displays correctly for both providers, and the Unsplash download endpoint is triggered on selection.
- Generating the same word for a second user makes zero LLM, image or TTS calls.
- All new UI strings are in vi and en.
