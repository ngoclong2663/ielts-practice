# Epic 10: Reading & Listening Practice

| Field | Value |
|---|---|
| Status | Phase 1 (Reading) built (2026-11-20) on `feat/epic10-reading-listening`: `exam` module, migration `20261120000019`, test bank, reading player with browser-only highlights, marking, results, review, redo, history. Set 1 complete for Reading: 4 tests, 12 passages, 160 questions with English explanations, all passing the seed checks. Phase 2 (Listening) built: free local TTS (Kokoro-82M + MeloTTS) via `scripts/exam-audio`, audio in S3, listening player (practice and exam modes with reading pauses), transcript review. Set 1 listening complete: 16 parts (4–6 min each, British, American and Australian voices, one map), audio in S3. Writing's "Inspired by" label removed; content workflow in the `ielts-content-seeding` skill. Interactive prototype: [IELTS Test Player](https://claude.ai/artifact/GGNUenrMcP7SrZsWqSMbcQ) (sample content written for it) |
| Depends on | Epic 1 (auth, i18n, shell, focus layout), Epic 5 (seed packs, S3 media), Epic 8 (content built locally with Claude Code, then reviewed) |

## 1. Goal

Learners practise IELTS Academic Reading and Listening on full tests or on one part at a time, then see why each answer is right.

1. **Practise by part.** Pick one passage (or one listening part), or any mix of parts from one test, instead of always sitting a 60-minute test. A band 5 learner can do one 20-minute passage on a bus; a full test is still one click away.
2. **Read like the computer test.** Passage and questions side by side, a question navigator, flags for review, a timer, and **highlighting that stays in the browser** (no API call, nothing stored on the server).
3. **Listen two ways.** *Practice mode*: pause, rewind 5 s, seek, speed 0.75–1.25×. *Exam mode*: plays once, with the reading pauses of the real test.
4. **Learn from every answer.** After submitting: the score (band for a full test), accuracy per question type, and for each question an English explanation, the **paraphrase pairs** between question and text, the common trap, and a button that shows the evidence in the passage or replays it in the audio.

Why these choices (from the reference survey, §12):
- Learners at band 4.5–6.5 lose most marks on **paraphrase** (the question and the text say the same thing in different words) and on **traps** (a figure that is mentioned and then corrected). An explanation that only says "paragraph C" does not teach that, so ours pairs the words.
- Short sessions get done. Practice by part is the most used mode on the Vietnamese sites we looked at; full tests are for the last weeks before the exam.
- The computer-delivered test is now the common format. Practising on the same layout (split screen, navigator, highlight) removes surprises on test day.

## 2. Scope

**In scope**
- Test bank for Reading and Listening, organised book → test → part.
- Practice by part (any subset of one test's parts) and full test, timed or untimed.
- Reading player: split view, resizable, text size, highlight in the browser, navigator, flags, timer, submit.
- Listening player: practice and exam modes, audio from S3, question layouts for forms, notes, tables, maps and multiple choice.
- Server-side grading, band for full tests, attempt history, accuracy by question type.
- Review: explanations, paraphrase pairs, traps, "show in passage", transcript with "listen again from here".
- "Redo the wrong ones" from the result page.
- Content for **one book first** (4 tests: 12 passages, 16 listening parts), then one book at a time.

**Out of scope (later)**
- General Training reading.
- Notes on the passage (the computer test has them; highlight comes first).
- Drill by question type across tests (the data supports it: every group has a type).
- Look up a selected word, or add it to a deck from the passage (Epic 3/8 integration).
- Speaking.
- Sharing results, leaderboards, teacher view.

## 3. Content source and copyright (decision needed first)

The practice books in `data-cambridge/` (Cambridge IELTS 16, 18, 19, 20, 21; PDF + audio, git-ignored) are the reference. Epic 5 §6.1 already set the rule for writing: **the books are inspiration only and are never copied**, and the scans never reach the repository, S3 or production. Reading passages and recorded audio are the books' main copyrighted content, so this epic needs the same decision made explicitly.

| | **A. Our own tests, modelled on each book test** (recommended) | **B. The book tests as published** |
|---|---|---|
| Passages and scripts | Written by us on the same broad topic, with our own facts, structure and sentences | Copied from the PDF |
| Questions | Same type sequence, counts, word limits and difficulty as the book test; our own wording | Copied |
| Audio | Our scripts recorded with multi-speaker TTS at build time (British, Australian and North American voices) | The book's MP3s |
| Label shown | None: tests are named by our own book and test number ("Book 21 · Test 1"); `source_ref` is kept internally only | "Cambridge IELTS 21 · Test 1" |
| Explanations | Ours (both options) | Ours |
| Fits Epic 5 §6.1 | Yes | No: rewrites the rule |
| Risk | Content effort; TTS is less natural than studio actors | Copyright claim against a public site; book audio and text on S3 |
| Extra benefit | We know every audio timestamp exactly (we build the audio from lines), so the transcript sync and "listen again" are exact | Learners can check against their own book |

The rest of this plan works for either option: the content format, players and review are the same. Only the build pipeline (§8) differs.

What the books give us under option A (checked 2026-11-20):
- **Blueprints.** Per test: topics, question-type order, group sizes, word limits, and where the answers sit in the text. Read from the page images.
- **Text layer.** Only cam21 has one (OCR, tables scrambled). cam16, cam18, cam19 and cam20 (official) are scans; `CAM 20 Reading - Listening (đẹp).pdf` is a clean retyped copy. Blueprints are read from page images either way.
- **Answer evidence in audioscripts.** Each answer is underlined and marked `Q1`…`Q40` in the margin. This shows how answers are placed in a script (the correction trap, the spelled name, the number said twice). We follow the same pattern when writing our scripts.
- **Audio.** cam16, cam18, cam19 and cam20 have one file per part. cam21 has one file per test. Under option A the book audio is used only to calibrate pace (about 150 words a minute) and accents.

## 4. Practice modes

| Mode | What | Timer | Score shown |
|---|---|---|---|
| **By part** | Any subset of one test's passages or parts, e.g. Passage 2 only, or Listening Parts 1 + 2 | Suggested (20 min per passage; listening runs as long as its audio), a custom time, or none (counts up) | `9/13` correct and accuracy per question type. An estimate for 40 questions, not a band |
| **Full test** | All 3 passages (40 questions, 60 min), or all 4 parts (40 questions, about 30 min + 2 min check) | Countdown; ends the test at 0 | Correct `/40` and **band** (§7.3) |

- When the suggested time runs out in a **part**, a notice appears ("Hết thời gian gợi ý…") and the learner may keep going. The result records the overrun. In a **full test** the time limit submits the test, as in the exam.
- **Listening practice mode** allows pause, −5 s, seek and speed, and can skip the reading pauses. **Exam mode** plays once with no pause or seek, inserts the reading pauses (§6.2), and in a full test adds the final 2-minute check. Exam mode is the default for a full test; practice mode for parts.
- **Redo the wrong ones.** The result page opens a new attempt with only the questions answered wrongly, in the same player. It is untimed, scored on its own, does not replace the first result, and is kept out of the history stats.

## 5. Reading player

### 5.1 Layout

Desktop and tablet (≥ 861 px) — see prototype screen "Làm Reading":

```
┌ ✕ Thoát │ Book 21 · Test 1 · Passage 1                     18:42  Aa  [Nộp bài] ┐
├──────────────────────────────────┬─┬────────────────────────────────────────────────┤
│ Passage (scrolls on its own)     │║│ Questions (scroll on their own)                │
│ title, subtitle                  │║│ Câu 1–4  TRUE / FALSE / NOT GIVEN              │
│ A  paragraph …                   │║│  1  statement …         [TRUE][FALSE][NOT GIVEN] ⚑ │
│ B  paragraph … ▒highlight▒ …     │║│ Câu 5–7  Notes with inline gaps                │
│                                  │║│  … reduced to [ 5 ______ ] per cent …          │
├──────────────────────────────────┴─┴────────────────────────────────────────────────┤
│ Passage 1 9/13 │ [1][2][3]…[13]   ‹ ›       (Passage 2 · Passage 3 tabs if selected) │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

- **Top bar:** exit (asks before leaving with unsent answers), the test and part, the timer (tap to hide; turns red and is announced at 10 and 5 min left, as on the computer test), text size and contrast (remembered per learner), submit.
- **Split view:** passage on the left at 55 % by default, resizable from 32 % to 72 % with the pointer or the arrow keys. The width is remembered. Paragraph letters sit in the margin. Passage text is 17 px (15 / 17 / 20), line height 1.65, at most 70 characters per line.
- **Navigator:** one tab per selected part with its answered count, a chip per question (answered = tinted, current = ring, flagged = dot), previous and next. Clicking a chip scrolls to the question and focuses its input.
- **Flags:** ⚑ on each question ("Đánh dấu xem lại"); the submit dialog lists unanswered and flagged questions.
- **Phone (< 861 px):** one pane at a time with a switcher under the top bar, `[Bài đọc] [Câu hỏi 9/13]`. Choosing a question in the navigator switches to the questions pane. The bottom bar becomes `‹  Câu 5 / 13  ›`, and tapping the middle opens the chip grid as a sheet.
- **Style:** the app's emerald brand in the `accent-ladder` style (taste skill). The full hue goes only on small marks (current question ring, chosen answer border, progress); pale tints carry the areas (answered chips, chosen options). The primary button is near-black. No gradients or shadows at rest; the selection toolbar and dialogs float.

### 5.2 Highlight (browser only)

- **Make one:** select text in the passage or in a question stem. A small dark toolbar appears above the selection (below it on touch screens, clear of the system menu): `Tô sáng · Huỷ`. Keyboard: ⌘/Ctrl + Shift + H.
- **Remove one:** click or tap inside a highlight, then choose `Bỏ tô sáng`. Text-size menu → `Xoá tất cả tô sáng` clears them all.
- **One colour** (yellow, readable in both themes). Overlapping highlights merge. A selection across paragraphs becomes one highlight per paragraph.
- **Rendering:** the CSS Custom Highlight API (`CSS.highlights` + `::highlight(user-hl)`). It paints ranges without changing the DOM, so React re-renders, answer inputs and text search are unaffected. Supported in Chrome/Edge 105+, Safari 17.2+ and Firefox 140+. In an older browser the toolbar does not appear and everything else works.
- **Storage:** each highlight is `{block, start, end}`: a stable block id (`p-B` for paragraph B, `q7` for a question stem) and character offsets in that block's text. The list sits in `localStorage` with the rest of the attempt state (§5.4). It survives a reload and is shown again on the review screen of that attempt. **It is never sent to the API.**
- The review screen paints the evidence with a second highlight (`::highlight(evidence)`: green tint with an underline), so the learner's own marks stay visible next to it.

### 5.3 Question types and how each is answered

Six building blocks cover every Academic Reading and Listening type:

| Block | IELTS types | Control |
|---|---|---|
| **Choice** | Multiple choice (one answer); TRUE/FALSE/NOT GIVEN; YES/NO/NOT GIVEN | Option rows with radio buttons; T/F/NG and Y/N/NG as three large buttons |
| **Choose N** | "Choose TWO / THREE letters" | Checkboxes. Choosing one more replaces the oldest. Counter "Đã chọn 1/2". Each correct letter scores one question, in either order |
| **Letter match** | Matching information (which paragraph), features (people), sentence endings, listening "choose from the box", map/plan labelling (letters on a picture) | A row of letter buttons when there are 8 options or fewer; a dropdown when there are more (headings i–x). The option list stays visible above the items |
| **Gap** | Sentence, note, table, flow-chart, form, summary and short-answer completion; diagram labelling | Text inputs placed inside the real layout (notes with bullets, a table with cells, a flow chart with arrows). A live word count turns red past the limit ("2 từ") |
| **Gap from a bank** | Summary completion with a word list (A–I) | A dropdown of letters per gap; the bank stays visible beside the summary |
| **Figure** | Maps, plans, diagrams | An image (our SVG) next to Letter match or Gap items; tap to zoom on a phone |

Matching headings, where the headings list is long, stays a dropdown per paragraph. Drag and drop was left out because it is slow on phones and hard to use from the keyboard.

### 5.4 Attempt state (browser only until submit)

- While a test is open, everything lives in `localStorage` under one key per run (`exam:<user>:<test>:<parts>:<mode>`): answers, flags, highlights, current question, and the timer as a **deadline timestamp** (a reload does not reset the clock).
- The test bank shows "Đang làm dở · Tiếp tục" on a run that has state. Starting again asks whether to discard it.
- On submit, only `{answers, duration, timer settings}` go to the API. Highlights move to a key per attempt id so the review screen can show them, and are pruned after 30 days.
- If storage is unavailable (private window), the run still works; it is just lost on reload.

## 6. Listening player

### 6.1 Layout

- **Audio bar** under the top bar: play/pause, −5 s, a seek bar with time, speed `0.75× · 1× · 1.25×` (practice mode). In exam mode: one play button that is disabled once started, a progress bar with no seek, and the line "Chỉ nghe một lần, không tạm dừng được".
- **Questions** in one centred column (760 px), with the navigator and part tabs at the bottom as in Reading. Forms and notes look like the paper: label column and value column with inline gaps.
- The listening part and the test are named in the top bar; the transcript is hidden until submit.
- **Phone:** the audio bar is sticky under the top bar, the questions scroll under it, inputs are 44 px tall.

### 6.2 Reading pauses (cues)

Each part's audio comes with a list of cues: `{at, seconds, label}` ("Look at questions 7–10", 30 s). In exam mode the player pauses the audio at each cue and shows a countdown band ("Đọc trước câu 7–8 · 0:10"), then resumes. In practice mode the cues are skipped unless the learner turns on "Giữ thời gian đọc đề". The audio files therefore hold no long silences, and the same file serves both modes.

### 6.3 Review

- Two columns: explanations on the left, the **transcript** on the right (prototype screen "Listening: transcript").
- The transcript is split into speaker lines with timestamps. The current line is tinted while playing; answer words are underlined and the line shows `Q4`.
- **Nghe lại đoạn này** on each question seeks to 2 s before the evidence line and plays. Clicking any transcript line plays from there.
- On a phone the transcript sits under the explanations, and "Nghe lại" scrolls to the line.

## 7. Results, review and grading

### 7.1 Result page

- **Score:** `9/13` (part) or `29/40 · Band 7.0` (full test) at 40 px, with the time taken against the suggested time.
- **The one finding** under the score: the weakest question type of this attempt ("Cần luyện thêm: TRUE / FALSE / NOT GIVEN · đúng 1/4").
- **By question type:** one row per type with correct/total and a bar (red under 50 %).
- **Answer table:** question, your answer (red or green), correct answer, type, `Giải thích`. "Over the word limit" is shown when that was the reason.
- Actions: `Xem lại với giải thích` (primary), `Làm lại N câu sai`, `Chọn passage khác`.

### 7.2 Explanation for one question

Shown in the review player beside the passage or transcript:
- Your answer and the correct answer (accepted variants joined with " / ").
- **Why** (English, 1–3 sentences, plain words a band 5 learner can follow; decided: explanations are English only, the UI around them stays vi/en).
- **Paraphrase pairs:** question words ≈ text words, e.g. `mainly to reduce costs ≈ mainly because LEDs use around half as much electricity`.
- **Trap** (when there is one): the distractor and why it is wrong, e.g. "£35 is the student price; she is not a student."
- **Xem vị trí trong bài · đoạn C** (Reading): scrolls the passage to the evidence and paints it. **Nghe lại đoạn này** (Listening).
- For NOT GIVEN: the closest sentence and what it does not say.
- **Báo lỗi đáp án** (report a wrong key or explanation): a short form that goes to the Epic 6 feedback board and Telegram, with the part, question and the learner's note attached. Wrong keys with no way to report them are the most common complaint on the sites we surveyed.

### 7.3 Grading rules (server side)

- Case and surrounding spaces are ignored; inner spaces are collapsed; curly and straight apostrophes are the same.
- Every accepted form is listed in the key: `["30", "thirty"]`, `["cafe", "café"]`, `["metal", "metals"]`, UK and US spellings. The key format allows optional words in brackets (`"(the) harbour"`), which expand at seed time.
- **Word limit:** an answer longer than the group's limit is wrong even if it contains the right word (`motion sensors` for ONE WORD ONLY; `in the morning` when the key is `morning`). A number counts as one word; a hyphenated word (`check-in`) counts as one. Contractions are never tested.
- Book-style key notation is accepted when we write keys: `/` separates alternatives (`sports grounds/sports halls`), brackets mark optional words (`10.25 (am)`), and the two combine (`sport(s)`).
- Spelling must match one of the accepted forms; there is no fuzzy matching.
- **Choose N:** each correct letter scores one of the group's questions, in any order; extra letters score nothing.
- TRUE/FALSE/NOT GIVEN and YES/NO/NOT GIVEN are answered with buttons, so there are no abbreviations to parse.
- **Band** (full tests only), from the published Academic tables:

| Listening raw | Band | | Academic Reading raw | Band |
|---|---|---|---|---|
| 39–40 | 9.0 | | 39–40 | 9.0 |
| 37–38 | 8.5 | | 37–38 | 8.5 |
| 35–36 | 8.0 | | 35–36 | 8.0 |
| 32–34 | 7.5 | | 33–34 | 7.5 |
| 30–31 | 7.0 | | 30–32 | 7.0 |
| 26–29 | 6.5 | | 27–29 | 6.5 |
| 23–25 | 6.0 | | 23–26 | 6.0 |
| 18–22 | 5.5 | | 19–22 | 5.5 |
| 16–17 | 5.0 | | 15–18 | 5.0 |
| 13–15 | 4.5 | | 13–14 | 4.5 |
| 10–12 | 4.0 | | 10–12 | 4.0 |

Below 10 the result says "dưới 4.0" rather than guessing a finer band. Sources differ on Listening 4.0 (10–12 or 11–12); we use 10–12. ielts.org publishes only the anchor points (16 = 5, 23 = 6, 30 = 7, 35 = 8) and says the cut-offs "vary slightly from test version to test version", so the table is stored as data and the result page says it is an estimate.

## 8. Content format and build

### 8.1 Seed files

One file per part, compiled into the binary like the writing packs:

```
be/seeds/exam/<book>/                 # e.g. cam21 (option A: "inspired by")
  book.json                           # title, source label, sort
  test1/reading-1.json … reading-3.json
  test1/listening-1.json … listening-4.json
  figures/*.svg                       # our own maps and diagrams
```

Audio is not embedded (about 3–4 MB per part as 64 kbps mono MP3). `api exam-media <dir>` uploads a book's audio to S3 under content-addressed keys (`exam/audio/<sha256>.mp3`) and prints the keys that go into the seed files. Without S3, the media endpoint serves files from `EXAM_MEDIA_DIR`; e2e uses a 3-second fixture.

```jsonc
// reading-1.json
{
  "slug": "cam21-t1-r1", "source_ref": "cam21-test1-reading1",   // internal, never shown
  "skill": "reading", "part": 1, "title": "Light after dark", "subtitle": "…",
  "topic": "environment",                       // Epic 4 topic ids
  "paragraphs": [{ "id": "A", "text": "…" }],   // listening: "lines": [{ "id": "l12", "speaker": "MAN", "text": "…", "at": 84.2 }]
  "audio": null,                                 // listening: { "key": "exam/audio/<sha>.mp3", "duration": 412.6, "cues": [{ "at": 9.8, "seconds": 30, "label": "questions 1–6" }] }
  "groups": [
    { "type": "tfng", "from": 1, "to": 4, "instructions": "…",
      "items": [{ "n": 1, "text": "Councils replaced sodium lamps with LEDs mainly to reduce costs." }] },
    { "type": "notes", "from": 5, "to": 7, "word_limit": 1, "allow_number": true,
      "title": "The Fulda scheme", "lines": ["After midnight, … at {5} per cent of full power."] },
    { "type": "choose_n", "from": 8, "to": 9, "pick": 2, "stem": "…", "options": [["A", "…"]] }
  ],
  "key": {                                       // never sent to the client before submit
    "1": { "answer": ["TRUE"],
           "evidence": { "block": "A", "quote": "mainly because LEDs use around half as much electricity" },
           "pairs": [["mainly to reduce costs", "mainly because LEDs use around half as much electricity"]],
           "why": "…",                                // English
           "trap": "…" }                              // optional
  }
}
```

Group `type` values (they are also the categories for accuracy per type): `tfng`, `ynng`, `mcq`, `choose_n`, `match_headings`, `match_info`, `match_features`, `match_endings`, `summary_bank`, `sentence`, `notes`, `table`, `flowchart`, `form`, `summary`, `short_answer`, `diagram`, `map`.

### 8.2 Checks run by `api seed` and by `TestCommittedExamSeeds`

- Numbering: Reading 1–40 across three passages (13/13/14 usually), Listening 1–40 across four parts (10 each). No gaps or duplicates.
- Every question has a key; letters exist among the options; T/F/NG values are valid; Choose N has exactly `pick` answers.
- Every gap answer obeys its group's word limit.
- Every evidence quote is found word for word in its paragraph or transcript line.
- Listening: line times increase and end before `duration`; every cue sits inside the audio.
- Every figure file exists and every group type is known.
- Idempotent with a `content_hash`, like the writing packs: unchanged parts are skipped, changed parts are updated in place (attempts keep pointing at them).

### 8.3 Building one book (local, like Epic 8)

Written by hand with Claude Code (not the app's AI keys), one test at a time, then reviewed:

1. **Blueprint.** Render the test's pages to images (PyMuPDF via `uv`, run in the scratchpad) and record per part: topic, question-type order, counts, word limits, how each answer is placed.
2. **Write** the passage or script, questions and key with explanations in the seed format (option A). Under option B: transcribe and check the book text instead.
3. **Listening audio** (option A): synthesise each speaker line with multi-speaker TTS, join the lines with natural gaps, and write the line times and cues into the seed file. Encode as 64 kbps mono MP3, then `api exam-media`. The TTS provider is chosen in a spike (Gemini TTS first, since its key is already configured; compare one part against one alternative for naturalness and accents).
4. **Validate** with `go test ./... -run TestCommittedExamSeeds`, then `make seed`.
5. **Review in the app:** the owner sits each part once in review mode (every explanation, every "show in passage") before the book is committed.

Effort: about 6,000 words of passage and script per test, plus 80 explanations. One book (4 tests) is the first milestone.

## 9. Data model

| Table | Columns |
|---|---|
| `exam_books` | id, slug (`cam21`), title (`Book 21`), sort, status |
| `exam_tests` | id, book_id, test_no, slug (`cam21-t1`) |
| `exam_parts` | id, test_id, skill (`reading` / `listening`), part_no, slug, title, topic, q_from, q_to, question_types `TEXT[]`, content JSONB (paragraphs or lines, groups; no key), audio JSONB, content_hash, status |
| `exam_keys` | part_id PK, key JSONB (answers, evidence, pairs, explanations). A separate table, so a list query can never leak it |
| `exam_attempts` | id, user_id, skill, test_id, part_ids `UUID[]`, kind (`part` / `full` / `redo`), parent_attempt_id, mode (`practice` / `exam`), time_limit_sec, duration_sec, answers JSONB, results JSONB (per question: given, correct, type), correct, total, band `NUMERIC(2,1)` NULL, created_at |

No table for in-progress runs (they stay in the browser, §5.4) and none for highlights.

## 10. Backend (`exam` module) and API

Module layout as in CLAUDE.md (`domain/`, `ports/`, `infrastructure/`, `internal/use-cases/` one per file), registered in `be/app/container.go`. Grading, word-limit counting and band conversion are pure functions in `domain/rules/` with table tests.

| Method & path | Purpose |
|---|---|
| `GET /api/v1/exam/books?skill=reading` | Books → tests → parts (title, question count, types), with the learner's best score per part |
| `GET /api/v1/exam/tests/:slug?skill=&parts=1,3` | The chosen parts' content for the player (no key) |
| `GET /api/v1/exam/media/:sha.:ext` | 302 to a presigned S3 URL (audio, figures), like writing figures and guide videos |
| `POST /api/v1/exam/attempts` | `{test, skill, parts[], kind, mode, time_limit_sec, duration_sec, answers, parent_attempt_id?}` → graded attempt id |
| `GET /api/v1/exam/attempts/:id` | Attempt with results and, for its parts only, the key and explanations |
| `GET /api/v1/exam/attempts?skill=&cursor=` | History |
| `GET /api/v1/exam/stats?skill=` | Accuracy per question type over the last 20 attempts, and the bands of full tests over time |

New error codes in `be/pkg/exceptions/exception.go`: `EXAM_PART_NOT_FOUND`, `EXAM_PARTS_FROM_DIFFERENT_TESTS`, `EXAM_ATTEMPT_NOT_FOUND`, `EXAM_ANSWERS_INVALID`, translated in `errors.*`.

## 11. Frontend

- **Navigation:** two new sidebar sections beside Vocabulary and Writing:
  - **Reading:** "Đề luyện" `/reading`, "Kết quả" `/reading/history`
  - **Listening:** "Đề luyện" `/listening`, "Kết quả" `/listening/history`

  The mobile section switcher gains both.
- **Test bank** (`(app)/reading`, `(app)/listening`). The style sheet's list pattern: a 24 px title, underline tabs per book, then one block per test. Each block lists its parts as rows: checkbox, part, title, question types, best score pill. Each test has a "Làm cả đề" button. A sticky start bar appears once a part is ticked: count, timer choice (Listening: mode choice too), and `Bắt đầu · 40 phút`.
- **Players and review** in the `(focus)` route group (no sidebar): `/reading/[test]/practice?parts=1,2`, `/reading/attempts/[id]` (review), and the same for listening.
- **Feature folder** `fe/src/features/exam/`, shared by both skills: `components/player/` (split view, navigator, timer, submit dialog), `components/questions/` (the six blocks of §5.3), `components/highlight/` (`useHighlights`, selection toolbar), `components/audio/` (player, cues), `components/review/`, `hooks/useRunState` (the `localStorage` run of §5.4).
- **History pages:** attempts list (date, parts, score/band, time) and accuracy per question type, sorted weakest first.
- Every string in `en.json` and `vi.json`. IELTS instruction lines stay in English, as in the exam.
- **Accessibility:** every control reachable by keyboard; T/F/NG buttons are a radio group; Alt + ←/→ moves between questions; the timer warnings use `aria-live`.

## 12. Reference products (UI survey, 2026-11-20)

Sources: the official computer-delivered IELTS (Inspera), ielts.org format and sample pages, study4.com, ieltsonlinetests.com (IOT) and engnovate. idp.com and the British Council pages could not be opened, so their details come from secondary guides.

| Seen in | Pattern | Where we use it |
|---|---|---|
| Official test | Split passage/questions, each scrolling on its own; a 1–40 bottom bar that marks answered questions; a "Review" flag; previous/next arrows | §5.1 |
| Official test | Select text → Highlight (or Notes), "Clear all"; also usable in Listening | §5.2 (highlight first, notes later) |
| Official test | Timer flashes at 10 and 5 min; three text sizes and contrast options | §5.1 |
| Official test | Listening plays once, with reading time before each section and 2 min to check at the end | §4, §6.2 |
| Official test | Drag and drop for flow-chart and summary from a list | Replaced by a dropdown per gap (§5.3) |
| study4 | "Luyện tập" with a part picker (parts tagged with their question types) and a time limit from a dropdown or none; band only for a full test | §4, §11 test bank |
| study4 | "Làm lại các câu sai": untimed, kept out of the average | §4 |
| study4 | Explanation per question (theirs in Vietnamese, ours in English); save a passage word as a flashcard | §7.2; flashcards later (§2) |
| study4 | Restores the run after a network failure | §5.4 |
| IOT | Per question: "Locate", "Explain", "Listen from here", "Report"; "Keywords in question" next to "Similar words in passage" | §7.2 (pairs, evidence, report), §6.3 |
| engnovate | Mistakes analysed by question type (paid) | §7.1, free |

Complaints we design against:
- Wrong answer keys with no way to report them: our review has "Báo lỗi đáp án", and the seed checks every key.
- A transcript that does not match the audio: under option A we build the audio from the transcript lines.
- Explanations behind a paywall.
- Mock bands that feel unrealistic: we show a band only for 40 questions and call it an estimate.
- Replay in practice that is not clearly labelled as such: the mode is named in the audio bar, and a full test defaults to exam mode.

Also avoided: explanations that only give a paragraph letter, and an auto-submit that discards work when time runs out in practice.

## 13. User stories

- **US-10.1 Choose what to practise.** As a learner, I pick one or more passages (or listening parts) of a test, or the full test, choose a time limit and start.
- **US-10.2 Read and answer.** As a learner, I read the passage beside the questions, resize the split, change the text size, answer every question type, flag questions and move with the navigator.
- **US-10.3 Highlight.** As a learner, I highlight words in the passage and the questions and remove them; they survive a reload and never leave my browser.
- **US-10.4 Listen.** As a learner, I listen in practice mode (pause, rewind, speed) or exam mode (once, with reading pauses), and answer while listening.
- **US-10.5 Keep my place.** As a learner, if I close the tab or lose connection, I can continue the run with my answers, flags, highlights and remaining time.
- **US-10.6 Results.** As a learner, I see my score (band for a full test), my accuracy per question type and my weakest type.
- **US-10.7 Explanations.** As a learner, I read why each answer is right, see the paraphrase pairs and the trap, and jump to the evidence in the passage or replay it in the audio. If a key looks wrong, I report it in one step.
- **US-10.8 Redo mistakes.** As a learner, I redo only the questions I got wrong.
- **US-10.9 History.** As a learner, I reopen past attempts and see how my bands and question types change.
- **US-10.10 Add a book.** As the developer, I add a book's seed files and audio, the seed checks catch mistakes, and deploying makes the book available.

## 14. Phases

| Phase | Delivers | Content |
|---|---|---|
| **1. Reading** | Data model, seed loader and checks, test bank, reading player with highlight, grading, results, review, redo, history | Book 1 reading: 12 passages |
| **2. Listening** | Audio build pipeline and TTS spike, media endpoint, listening player (both modes, cues), transcript review | Book 1 listening: 16 parts |
| **3. More books and stats** | History stats by question type; then later items from §2 | One book at a time: cam20, cam19, cam18, cam16 |

### As built (Phase 1)

- **Sets, not books.** Content is grouped in sets named by us (`s1` = "Set 1" / "Bộ đề 1"); the reference book is only `source_ref` in the seed (`cam21`), never returned by the API.
- **Seed:** `be/seeds/exam/<set>/book.json` and `test<N>/reading-<P>.json`. `LoadExamSeeds` checks everything before writing (also run by `TestCommittedExamSeeds`); unknown JSON fields are refused.
- **API:** as §10, except `GET /exam/tests/:slug?skill=` returns every part of the skill (the player picks the chosen ones) and `POST /exam/attempts` takes `parts` as part numbers. The transcript (`lines`) is removed from the player content and sent only with the key.
- **Routes:** bank `/reading`, history `/reading/history` (app shell); player `/reading/practice/<test>?parts=1,2&time=2400` (`?kind=full`, `?redo=<attempt>&q=6,9`), result `/reading/attempts/<id>`, review `/reading/attempts/<id>/review?q=5` (focus layout).
- **Navigation:** a third sidebar section, Reading; the mobile switch cell cycles through the sections. Listening joins in Phase 2.
- **Test bank:** each test is a collapsible block (header keeps "Full test" and a done count). By default the first unfinished test opens, plus any with a run in progress or ticked parts; once the learner toggles one, the open set is remembered per skill in the browser.
- **Highlights:** `features/exam/lib/highlights.ts` + `hooks/use-highlights.ts` (CSS Custom Highlight API); the run state is `exam.run.*` in localStorage and highlights move to `exam.hl.<attempt>` on submit. The e2e test asserts that no API request carries them.
- **Tests:** rule tables (`domain/rules/rules_test.go`), committed seeds, integration `tests/integration/exam_test.go`, Playwright `fe/e2e/reading.spec.ts` (desktop and mobile).

### As built (Phase 2)

- **Audio build:** `scripts/exam-audio/` (`setup.sh`, `build.py`, one worker per engine). Each transcript line is rendered with its speaker's voice from the part's `voices` map (`kokoro:bf_emma`, `melo:EN-AU`), cached by text, joined with natural pauses, encoded as 64 kbps mono MP3 named by its hash; the build writes line times, cue times (`after` a line), duration and `audio.key` back into the seed file. `say` gives a spoken form when the transcript differs (spelled names).
- **Upload:** `api exam-media <file.mp3>` → S3 `exam/media/<sha256>.mp3`. `GET /api/v1/exam/media/:file` (public, like guide videos: the names are hashes and `<audio>` cannot refresh a session) redirects to a presigned link.
- **Figures:** map/plan SVGs in `be/seeds/exam/<set>/figures/`, named by hash at seed time and served from the binary at `/api/v1/exam/figures/:file`.
- **Player:** `ListeningPlayer` + `use-listening-audio`: practice (pause, −5 s, seek, 0.75–1.25×, optional reading pauses) and exam (one start, no pause or seek, stops at each cue with a countdown, moves to the next part, full test ends with a 2-minute check and submits). Audio position, cues taken and the mode are kept in the run state, so a reload resumes.
- **Review:** explanations beside the transcript; the spoken line is tinted, answer words underlined with `Q` markers, "Listen again · 0:42" plays from just before the evidence.
- **Single-file check:** `EXAM_PART=exam/<set>/test<N>/<skill>-<P>.json go test ./app/modules/exam/internal/use-cases/ -run TestExamSeedFile -v`.
- **Skill:** `.agents/skills/ielts-content-seeding/` documents the whole content workflow (reading, listening, writing) with PDF and scan scripts.

## 15. Decisions

Decided (2026-11-20):
1. **Content source: option A.** Our own tests modelled on each book test (§3), with **no "Inspired by" label**: books and tests carry our own names; `source_ref` stays internal.
2. **First book: cam21.**
3. **Explanations in English only.** The UI around them stays in vi/en.
4. **Navigation: two sidebar sections**, Reading and Listening (§11).
5. **Audio goes to our S3** (content-addressed keys, served through the media endpoint).

6. **Listening audio: free, local text-to-speech** built from our scripts (2026-11-20): Kokoro-82M (Apache 2.0; American and British voices) and MeloTTS (MIT; the Australian voice), run on the owner's Mac by `scripts/exam-audio/build.py`. Gemini TTS was the first proposal and stays a fallback.
7. **No reference-book labels anywhere**, writing included: the "Inspired by Cambridge IELTS…" label was removed from writing tasks.

Defaults in force:
- A band is shown only for a full test.
- In-progress runs and highlights stay in the browser; only submitted attempts reach the server.
- Explanations are written once with Claude Code and reviewed by the owner (Epic 8 approach), with no AI calls in production.
- A suggested timer that runs out in part practice warns but does not submit.

## 16. Definition of done

- Book 1 seeded in production: Reading in Phase 1, Listening in Phase 2. Audio served from S3.
- A learner can do a single passage, a mix of parts and a full test in both skills, on desktop and on a phone, in both languages.
- Highlights work in Chrome, Safari and Firefox, survive a reload and are absent from every API request (checked in e2e).
- Grading unit tests cover word limits, variants, Choose N and band boundaries; the seed checks run in `make test`.
- Playwright e2e: one part with highlight, submit and review; one listening part in exam mode with a cue.
