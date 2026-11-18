# Epic 7: Useful Vocabulary, AI Cost and Guide Videos

| Field | Value |
|---|---|
| Status | In progress (2026-11-18) on branch `feat/epic7-useful-vocab`: useful vocabulary (§3) and AI cost (§4) built; guide videos (§5): two videos published and shown in the app |
| Depends on | Epic 3 (card generation, LLM chain), Epic 5 (writing lessons), Epic 6 (admin board) |

## 1. Goal

1. **Useful vocabulary in writing.** Learner feedback (2026-11-18, on a Task 1 result page): *"Please add more useful vocab related to the topic."* The Epic 5 "Useful language" list had 6–10 bare phrases with a Vietnamese meaning. Learners need the words of the topic **and the other ways to say them**, so they can paraphrase the question and stop repeating words (IELTS Lexical Resource).
2. **AI cost on the admin board.** What Claude and Gemini cost, per model and per feature, including calls that failed.
3. **Guide videos.** Short narrated videos that show how to learn vocabulary Anki-style, stored on S3 and listed in the app.

## 2. Scope

**In scope**
- A "Vocabulary" step in every writing attempt (levels 1–3 while writing, every level after submitting) and on the result page.
- Words go to a deck through the Epic 3 flow (review the generated cards, choose the deck, save).
- A ledger of every model call; the admin board prices it.
- Guide videos made with HyperFrames (HTML rendered to MP4), narrated in Vietnamese.

**Out of scope (later)**
- Vocabulary for learners' own topics, or outside writing lessons.
- Marking which vocabulary words the learner used in their answer.
- Spend alerts or limits; costs per user.
- Videos in English; videos made by learners.

## 3. Useful vocabulary (built)

### 3.1 Content

Written once per lesson by the AI coach (Claude first, Gemini fallback), stored in `writing_lesson_vocab` and shared by every learner of the lesson. Lessons written before Epic 7 get theirs on first request, so no lesson is regenerated.

| Group | What | Count |
|---|---|---|
| `topic` | Words and collocations a band-7 answer on this topic uses ("labour market", "manufacturing sector") | 7–9 |
| `paraphrase` | Key words of the question, with the alternatives to restate them in the introduction ("show" → "illustrate", "depict") | 5–7 |
| `task` | Language the task type needs: Task 1 change and comparison (rise, overtake, account for); Task 2 opinion, cause, effect, concession | 5–7 |

Each item: `term`, `pos`, `vi`, `synonyms` (2–4, same form as the term and interchangeable in the example), `example_en` (new sentence on the topic; no numbers in Task 1 examples so nothing contradicts the figure), `note_vi` (difference between synonyms, or a common mistake). The lesson's model answer is passed to the coach so the vocabulary matches it, but examples are never copied from it.

The lesson no longer asks for `useful_language` (stored lessons keep theirs, unused).

### 3.2 When learners see it

| Level | While writing | After submitting |
|---|---|---|
| 1–2 | Vocabulary step + compact reminder in the final step | Result page |
| 3 | Vocabulary step + compact reminder | Result page |
| 4 (free writing) | Hidden (`WRITING_VOCAB_LOCKED`) | Result page |

Rule: `rules.VocabVisible(level, submitted)`.

### 3.3 Saving words

The learner ticks words and presses **Add N words to a deck**. The app creates an Epic 3 generation job (`POST /vocab/jobs`) with each term and its Vietnamese meaning as the sense hint, then opens `/add-words?job=<id>` to review the cards and pick the deck. Card generation limits apply as usual.

### 3.4 API

| Method & path | Purpose |
|---|---|
| `GET /api/v1/writing/attempts/:id` | Adds `vocab_available`; drops `useful_language` |
| `GET /api/v1/writing/attempts/:id/vocab` | `{groups: [{key, items: [...]}]}`; writes the lesson's vocabulary on first request (one AI call shared by everyone; concurrent requests wait for it); rate limited |

Migration `20261118000012`: table `writing_lesson_vocab (lesson_id PK, content, model, tokens)`; `writing_ai_requests.kind` gains `vocab`.

## 4. AI cost on the admin board (built)

- **Ledger.** The llm module wraps each model of the chain and writes one `ai_usage` row per call: provider (`anthropic` / `google`), model, feature (`vocab.cards`, `vocab.topic_words`, `writing.lesson`, `writing.vocab`, `writing.check`, `writing.submit`), ok, input / output / cache-read tokens. Failed calls are recorded too: a fallback answer also pays for the failed first call, and a model that keeps failing (for example a 429) shows up. Gemini's prompt count includes cached tokens; the ledger stores them apart, as Claude reports them.
- **Prices.** USD per million tokens in `admin/domain/rules/ai-cost.go` (checked 2026-11-18): Claude Haiku 5.5 $0.10 in / $0.50 out (cache read $0.01), Gemini 3.1 Flash-Lite $0.25 in / $1.50 out (cache $0.025), plus the other current Claude models. Costs are computed on request, so a price change reprices the history. A model without a price shows "?".
- **Board.** Overview → "AI cost (estimated)": all models and Claude / Gemini for 30 days and all time; a table per model (price, calls, failed calls, tokens, cost); cost per feature.
- **History.** Migration `20261118000013` copies past writing calls (which recorded their model). Card generation and topic lists did not record a model, so they count from this release.

## 5. Guide videos

### 5.1 Made so far (for review)

| Slug | Title | Length | Where in the app |
|---|---|---|---|
| `srs-basics` | Học từ vựng kiểu Anki | ~52 s | Vocabulary (decks, study) |
| `writing-levels` | 4 mức hỗ trợ khi luyện viết | ~50 s | Writing (level picker on the task page) |

- `srs-basics` follows [Hướng dẫn sử dụng Anki](https://vumt.edu.vn/huong-dan-su-dung-anki): open the deck and due cards come one by one; two-sided cards (word; meaning, phonetics, example, image); recall before showing the answer; rate honestly (Lại = quên, Khó, Tốt, Dễ); the app schedules the next review from the rating (forgotten cards return more often, known cards are spaced further); review every day, 15–20 minutes. It shows no interval numbers, because the guide gives none and they depend on each card.
- Made with HyperFrames: one HTML scene per narration line, every reveal timed from the voice's word timings; Vietnamese captions; the app's colours, font (Be Vietnam Pro) and UI pieces redrawn in HTML.
- `writing-levels` explains the four support levels of Epic 5 §4 (Hỗ trợ đầy đủ → Ghi chú ý chính → Dàn ý tiếng Anh → Như thi thật) with the real first sentence of the cam21 Task 1 lesson at each level, then sentence feedback, the estimated band and "Thử mức …": support fades until the learner writes the whole answer.
- Voice: VieNeu V4 "Mai Bé Phương" (a native southern female voice), made through the VieNeu MCP tool on the user's VieNeu account; word timings come from the audio's pauses matched to the script's punctuation (`align.mjs`), and each line is loudness-normalised. Music bed and sound effects from the HeyGen catalog.
- Sources stay local (`scripts/guide-videos/`, git-ignored): scripts, voice files and the generator. Only the final MP4 is published.

### 5.2 Publishing (built)

- The approved MP4 (H.264 1080p, `faststart` so playback starts at once) and a poster frame (JPEG) go to S3 under content-addressed keys, `guides/videos/<sha256>.mp4` and `guides/posters/<sha256>.jpg`, with `api upload-media <file>...` (it prints the names for the seed file).
- `be/seeds/guides/videos.json` (committed): slug, topic (`vocab` | `writing`), sort, narration language, title and description in vi/en, duration, video and poster names. `api seed` loads it into `guide_videos` (migration `20261118000014`), idempotent like the task packs.
- API: `GET /api/v1/guides?topic=` (signed in) lists published guides; `GET /api/v1/guides/media/<sha256>.mp4|jpg` (public, like figures) redirects to a 2-hour presigned link, cached 90 minutes, and only for files a published guide uses.

### 5.3 In the app (built)

| Where | What |
|---|---|
| Decks page | `srs-basics` as a card: poster with a play button, title, description, "Xem · 0:52" and "Để sau" |
| Writing Task 1 / Task 2 lists | `writing-levels` as the same card |
| Task page, level picker | "Chưa biết chọn mức nào? Xem video 0:50" link, where the learner makes the choice the video explains |

- The video plays in a dialog (native controls, Vietnamese captions burned in). The English UI adds "Narrated in Vietnamese, with Vietnamese captions."
- Once 80% has played, or the learner presses "Để sau", the card shrinks to a one-line "Xem lại: …" link. This memory is per browser (localStorage); losing it only shows the card again.

## 6. Open questions

1. Videos in English for the English UI, or Vietnamese only.
2. Claude never answered: `ANTHROPIC_API_KEY` held a Claude subscription token (`sk-ant-oat…`), which the Messages API rejects, so Gemini answered every call. Fix: an API key from console.anthropic.com (`sk-ant-api03-…`) locally and in production. The llm module now refuses a subscription token at start-up with a clear error and skips Claude instead of failing on every call.
