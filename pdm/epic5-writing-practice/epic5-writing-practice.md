# Epic 5: Guided Writing Practice (IELTS Writing Task 1 & 2)

| Field | Value |
|---|---|
| Status | Planned. The task bank and its seed (§6, §7) are implemented on branch `feat/epic5-writing`; the practice flow is not built yet. |
| Depends on | Epic 1 (auth, i18n, shell), Epic 3 (LLM providers, S3/storage, quotas pattern) |

## 1. Goal

Help **weak learners** (around band 4.0–5.5) write a full IELTS answer. Many of them cannot yet organise ideas, even in Vietnamese, and cannot turn those ideas into correct English. The app walks them through one task **step by step**:

1. Understand the task and choose ideas (in Vietnamese).
2. Build each paragraph sentence by sentence: read a Vietnamese sentence, write it in English, and get feedback on meaning, grammar and word choice.
3. Write the full answer from their own sentences, then get feedback on the four IELTS criteria and an estimated band range.

Support is reduced over time ("I do → we do → you do"), so learners end up writing without Vietnamese prompts.

Why this design (research notes, 2026-11-15):
- Low-proficiency writers plan better in their first language: they produce more ideas and more detail (Jones & Tetroe 1987; later studies of ESL/EFL learners agree).
- Most errors of Vietnamese learners come from word-for-word transfer ("drink medicine", dropped subjects). Feedback must check **meaning**, accept any correct English and name the literal-translation mistake, never require one exact answer.
- Support must fade. If learners always get full Vietnamese sentences, they learn to translate, not to write.
- AI band scores agree with examiners only moderately (often within about one band), so the score is shown as an **estimated range** and the comments come first.

## 2. Scope

**In scope**
- Navigation grouped into sections: **Vocabulary** and **Writing**.
- A task bank of Academic Task 1 and Task 2 prompts, seeded pack by pack (§6).
- Guided practice for one task at four support levels (§4).
- Sentence and paragraph feedback, plus full-answer feedback with an estimated band.
- Practice history: attempts, answers and feedback.
- Daily limits for AI feedback.

**Out of scope (later)**
- General Training Task 1 (letters).
- Learners typing in their own prompts or uploading images.
- Human (teacher) marking, sharing essays, leaderboards.
- Speaking, reading, listening.

## 3. Navigation

The sidebar becomes collapsible sections (an accordion). Each section remembers whether it is open (local storage), and the section that holds the current page is always open.

```
Dashboard
▾ Vocabulary
    Decks
    Add words
▾ Writing
    Task 1 (charts, maps, processes)
    Task 2 (essays)
    My writing (history)
Settings
```

- **Mobile:** the bottom bar shows the current section's items, plus a section switcher (Vocabulary / Writing). An accordion does not fit a bottom bar.
- Routes: `/writing/task-1`, `/writing/task-2`, `/writing/tasks/:id`, `/writing/attempts/:id`, `/writing/history`.

## 4. Practice flow

```
Pick a task (filter: Task 1 type / Task 2 type / topic)
        │
        ▼
Choose support level      1  Full Vietnamese sentences + hints   (default for new learners)
                          2  Vietnamese idea notes only
                          3  English outline only
                          4  Just the question (free writing)
        │
        ▼
Step 0  Understand & plan (Vietnamese)
        Task 1: "What does the figure show?" / "Which 2–3 features matter most?"  (multiple choice)
        Task 2: question type, then your position and 2 reasons from options     (or write your own)
        │
        ▼
Step 1…n  One paragraph at a time
        Task 1: Introduction → Overview → Body 1 → Body 2
        Task 2: Introduction (paraphrase + position) → Body 1 → Body 2 → Conclusion
        For each sentence: Vietnamese prompt (level 1) or idea note (level 2) or outline (level 3)
          + optional hints (key collocation, sentence pattern), the learner writes English
        [Check paragraph] → per-sentence feedback: ✓ correct / ~ understandable / ✗ wrong meaning,
          the error type (literal translation, grammar, word choice, spelling), a short fix
          and one model sentence. The learner may rewrite before moving on.
        │
        ▼
Final step  Write the whole answer
        The learner's corrected sentences are shown as a draft (level 1–2) or hidden (level 3–4).
        Word counter (150 / 250) and an optional timer (20 / 40 min).
        [Submit] → feedback on Task Achievement / Task Response, Coherence & Cohesion,
          Lexical Resource, Grammatical Range & Accuracy; an estimated band range
          (e.g. 5.5–6.0); the 3 most useful fixes; a model answer.
        Useful collocations from the feedback can be saved to a vocabulary deck (Epic 3 flow).
```

Rules:
- One sentence prompt per sentence of a model answer. Task 1 ≈ 8–10 sentences, Task 2 ≈ 12–15.
- A learner can skip a step, but the final step is always available.
- An attempt can be resumed. Answers are saved when each paragraph is checked.
- The level suggested for the next task goes up after two finished attempts in a row with mostly ✓ sentences. The learner can always change it.

## 5. Content per task: the lesson

The guided steps come from a **lesson**: model answer, sentence prompts, hints and idea options for one task. It is generated once per task (and per Task 2 position), stored, and shared by all learners, like the Epic 3 lexeme cache.

```jsonc
{
  "plan": {                                   // Step 0
    "questions": [{ "q_vi": "...", "options_vi": ["..."], "answer": 1 }],      // Task 1
    "positions": [{ "id": "agree", "label_vi": "Đồng ý", "reasons_vi": ["..."] }]  // Task 2
  },
  "paragraphs": [
    {
      "key": "introduction",                  // introduction | overview | body1 | body2 | conclusion
      "sentences": [
        {
          "vi": "Biểu đồ đường thể hiện số lượng việc làm …",   // level 1
          "note_vi": "giới thiệu: biểu đồ gì, ở đâu, khi nào", // level 2
          "outline_en": "chart shows / jobs / four industries / 1965–2025", // level 3
          "hints": [{ "en": "employment in", "vi": "việc làm trong ngành" }],
          "model_en": "The line graph illustrates how many people were employed …"
        }
      ]
    }
  ],
  "model_answer": "…",
  "useful_language": [{ "en": "rise steadily", "vi": "tăng đều" }]
}
```

- Task 2 lessons are generated per position (`agree`, `disagree`, `balanced`), only when a learner first picks that position.
- Each lesson is checked against the task facts (§6.2), so numbers in the model answer match the figure.
- Bad lessons can be reported, as cards can be in Epic 3; reports send the lesson for regeneration.

## 6. Task bank and seeding

### 6.1 Sources and copyright

- The practice books we own (Cambridge IELTS 16–21) are **only a source of inspiration**. They are **never copied** into the app:
  - Every prompt is **rewritten in our own words**.
  - Every Task 1 figure is **redrawn by us** (our own SVG), with **changed data and labels**: different numbers, years, categories or layout, while keeping the same task type and difficulty.
  - Task 2 keeps the topic (ideas are not protected) but uses our own wording.
  - Book sample answers and examiner comments are **not used**. Model answers come from our lessons (§5).
- `source_ref` (for example `cam21-test1-task1`) records which book task inspired ours. Learners see it as a label, **"Inspired by Cambridge IELTS 21 · Test 1"** (decided), so they can find the original in their own book. It is never presented as the original wording.
- The scans and PDFs stay in `data-cambridge/` (git-ignored) and never reach the repository, S3 or production.

### 6.2 Seed packs

Seeds are versioned in the repository and **compiled into the API binary** (`go:embed`), so every environment gets the same tasks.

```
be/seeds/writing/
  cam21/
    pack.json            # the tasks, in our words
    figures/*.svg        # our own redrawn figures
  cam20/ …               # added one book at a time
```

`pack.json`:
```jsonc
{
  "pack": "cam21",
  "tasks": [
    {
      "slug": "cam21-test1-task1",          // stable id, unique across packs
      "source_ref": "cam21-test1-task1",    // internal
      "test_no": 1, "task_no": 1,
      "task_type": "line_graph",            // §6.3
      "topic": "work",                      // Epic 4 topic ids
      "title": "Jobs in four industries, 1965–2025",
      "prompt": "The line graph shows …",
      "figures": [{ "file": "figures/test1-task1.svg", "alt": "Line graph of …" }],
      "facts": { … }                        // the figure as data, for lessons and feedback
    }
  ]
}
```

- `facts` holds what the figure shows as data: chart series, table cells, process stages or map changes. Lessons and feedback use it to check numbers and details, so the AI never has to read the image.
- `min_words` and `minutes` follow from `task_no` (150 / 20 and 250 / 40).
- `prompt` holds only the task-specific text. The standard exam lines ("Summarise the information by selecting and reporting the main features…", "Give reasons for your answer…", "Write at least 150 words.") come from the UI catalogues, so every task shows them the same way.
- Figures are drawn as SVG on a white background (system sans-serif font, no external resources). The generator for a pack is a throwaway script; the SVG files in the repository are the source of truth and can be edited by hand.

### 6.3 Task types

| Task 1 (Academic) | Task 2 |
|---|---|
| `line_graph`, `bar_chart`, `pie_chart`, `table`, `mixed`, `map`, `process` | `opinion` (agree/disagree), `discussion` (both views + opinion), `advantages_disadvantages`, `problem_solution`, `two_part` (direct questions) |

### 6.4 Seeding rules

- Command: `api seed` (local: `make seed`, also run by `make be`). In production the container entrypoint runs it right after the migrations, so **a deploy that adds a pack seeds it automatically**. `RUN_SEEDS=false` turns this off.
- **Idempotent.** Each task stores a `content_hash` (SHA-256 of its JSON and figure bytes). An unchanged task is skipped; a changed task is updated in place (same id, so attempts keep their task); a new slug is inserted.
- Tasks are never deleted by the seed. A task removed from a pack stays in the database.
- **Figures go to S3** under content-addressed keys: `writing/figures/<sha256>.svg`. The same file is uploaded once and never changes, so dev and production can share a bucket safely. The bucket stays private, so the database stores the object **key**, not a URL (a presigned URL expires after an hour). The API serves figures through `GET /api/v1/writing/figures/<sha256>.svg`: a 302 to a fresh presigned URL, cached by the browser for 50 minutes. The task API returns that path as the figure URL.
- Production requires S3: seeding fails if a pack has figures and S3 is not configured. In development without S3 the task is stored and a warning is logged.

## 7. Data model

Implemented now:

```sql
writing_tasks (
  id            UUID PK,
  slug          VARCHAR(64) UNIQUE,      -- 'cam21-test1-task1'
  pack          VARCHAR(32),             -- 'cam21'
  source_ref    VARCHAR(64),             -- shown as "Inspired by Cambridge IELTS 21 · Test 1"
  test_no       SMALLINT, task_no SMALLINT (1|2),
  module        VARCHAR(16) 'academic',
  task_type     VARCHAR(32),             -- §6.3
  topic         VARCHAR(32),
  title         VARCHAR(160),
  prompt        TEXT,
  figures       JSONB  -- [{key, sha256, content_type, width, height, alt}]
  facts         JSONB,
  min_words     SMALLINT, minutes SMALLINT,
  status        VARCHAR(16) 'published' | 'draft' | 'archived',
  content_hash  CHAR(64),
  created_at, updated_at
)
```

Planned with the practice flow:

| Table | Purpose |
|---|---|
| `writing_lessons` | One per task (and per Task 2 position): `task_id`, `position`, `content` JSONB (§5), `status`, `model`, token counts, `report_count`. Unique on `(task_id, position)`. |
| `writing_attempts` | A learner's run through one task: `user_id`, `task_id`, `lesson_id`, `level` (1–4), `position`, `plan` JSONB (Step 0 answers), `status` (in_progress / submitted), `current_step`, timestamps. |
| `writing_attempt_paragraphs` | One row per checked paragraph: `attempt_id`, `paragraph_key`, `answers` JSONB (the sentences written), `feedback` JSONB, `checked_count`. Unique on `(attempt_id, paragraph_key)`. |
| `writing_submissions` | The final answer: `attempt_id`, `text`, `word_count`, `seconds_spent`, `feedback` JSONB (four criteria, band range, fixes, model answer), token counts. |
| `writing_ai_requests` | One row per AI feedback call (paragraph check or full answer) for the daily limits and cost tracking. |

## 8. Backend (`writing` module)

### 8.1 API

| Method & path | Purpose |
|---|---|
| `GET /api/v1/writing/tasks?task_no=&type=&topic=` | Task list (published only), with "done" flags for the user |
| `GET /api/v1/writing/tasks/:id` | Task with figure URLs and the "Inspired by …" label |
| `GET /api/v1/writing/figures/:sha.:ext` | Public; 302 to a presigned S3 URL (implemented) |
| `POST /api/v1/writing/tasks/:id/attempts` | Start an attempt `{level, position?}`; creates the lesson if missing |
| `GET /api/v1/writing/attempts/:id` | Attempt + lesson view for its level (model sentences hidden until checked) |
| `POST /api/v1/writing/attempts/:id/paragraphs/:key/check` | `{answers: [..]}` → per-sentence feedback (one AI call per paragraph) |
| `POST /api/v1/writing/attempts/:id/submit` | `{text, seconds_spent}` → full feedback |
| `GET /api/v1/writing/attempts` | History |
| `GET /api/v1/writing/status` | Daily limits left |

### 8.2 AI

- Same providers and models as Epic 3 for everything, lessons included (decided): `claude-haiku-5-5` first, `gemini-3.1-flash-lite` fallback, JSON-schema output. A stronger lesson model can be tried later if lesson quality is not good enough.
- **Lesson generation:** task prompt + `facts` + type → lesson JSON (§5). The model answer targets band 7, using simple, clear language a band-5 learner can copy.
- **Paragraph check:** Vietnamese source sentences + model sentences + the learner's sentences → per-sentence verdict, error type, short fix (Vietnamese explanation, English correction). Accepts any correct paraphrase.
- **Full answer:** public IELTS band descriptors (summarised in the system prompt) + task + facts → per-criterion comments, band range, top 3 fixes.
- Learner text is data, never instructions.

### 8.3 Limits (decided)

- `WRITING_DAILY_CHECKS` = 40 paragraph checks per user per 24 h.
- `WRITING_DAILY_SUBMISSIONS` = 5 full-answer feedbacks per user per 24 h.
- Lessons are shared, so generating them does not count against the user.

## 9. Web

- `features/writing/`: task list (filters by type and topic, "done" badge), task page (prompt, figure, level picker), the stepper (Step 0, paragraphs, final), feedback views and history.
- Figures use `<img>` from the figures endpoint, on a white background so charts read well in dark mode.
- The paragraph editor keeps one input per sentence; Enter moves to the next sentence.
- All copy in vi/en catalogues.

## 10. User stories

- **US-5.1 Navigation sections.** As a learner, I see Vocabulary and Writing as separate, collapsible sections, so I can find each practice area quickly.
- **US-5.2 Task bank.** As a learner, I can browse Task 1 and Task 2 tasks by type and topic and see which ones I have done.
- **US-5.3 Plan in Vietnamese.** As a weak learner, I answer simple Vietnamese questions about the task and pick my ideas before writing.
- **US-5.4 Sentence by sentence.** As a learner, I translate each Vietnamese sentence of a paragraph into English and get clear feedback that accepts any correct answer and explains literal-translation mistakes.
- **US-5.5 Full answer.** As a learner, I write the whole answer from my sentences and get criterion feedback, an estimated band range and a model answer.
- **US-5.6 Less help over time.** As a learner, I can pick a support level, and the app suggests a harder level when I am ready.
- **US-5.7 History.** As a learner, I can reopen past attempts and their feedback.
- **US-5.8 Seed packs.** As the developer, I add a pack of rewritten tasks and redrawn figures to the repository, and deploying it makes the tasks available in production.

## 11. Decisions

Decided (2026-11-15):
- Book material is inspiration only: prompts are rewritten and figures redrawn with changed data; book images, text and sample answers are not used (§6.1).
- Seeding is per book pack, starting with `cam21`; figures are stored in our S3 bucket (§6.4).
- Seeds ship inside the binary and run on every deploy after migrations, so production follows automatically.

- Academic only for v1; General Training letters later.
- Learners see an "Inspired by Cambridge IELTS 21 · Test 1" label built from `source_ref`.
- Daily limits as in §8.3: 40 paragraph checks and 5 full-answer feedbacks per user per 24 h.
- Lessons and feedback use the Epic 3 models: Haiku 5.5 first, Gemini 3.1 Flash-Lite fallback.

Open: none.

## 12. Definition of done

- Sidebar sections and mobile switcher work in both languages.
- At least the `cam21` pack is seeded in production, figures load from S3.
- A learner can complete a Task 1 and a Task 2 attempt at level 1 and level 4, with feedback.
- Daily limits enforced; AI failures show a retry, never lose the learner's text.
- Backend integration tests (seed idempotency, attempt flow with fake AI) and Playwright e2e for one full attempt.
