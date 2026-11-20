---
name: ielts-content-seeding
description: Add new IELTS practice content to this app — Reading passages, Listening parts (with locally generated TTS audio uploaded to S3) and Writing tasks — modelled on the practice books in data-cambridge/ but written in our own words. Use when asked to seed, add or build a new set/book/test of reading, listening or writing content, to extend an existing set, or to fix seeded content.
---

# Seeding IELTS content (Reading, Listening, Writing)

The practice books in `data-cambridge/` (Cambridge IELTS PDFs and audio, git-ignored) are **references only**. We read
them for their *blueprint* — topics, question types, counts, word limits, how answers are placed — and write
**our own** passages, scripts, tasks and explanations. Nothing from a book is copied into the repo, S3 or the app,
and no reference-book label is shown to learners (`source_ref` is internal). Read `references/sources.md` first.

## Where content lives

| Skill | Seed files (compiled into the API, loaded by `make seed` / every deploy) | Checks |
|---|---|---|
| Reading | `be/seeds/exam/<set>/test<N>/reading-<P>.json` (+ `book.json` per set) | `go test ./app/modules/exam/...` |
| Listening | `be/seeds/exam/<set>/test<N>/listening-<P>.json` (+ audio in S3, figures in `be/seeds/exam/<set>/figures/`) | same, after `scripts/exam-audio/build.py` |
| Writing | `be/seeds/writing/<pack>/pack.json`, `figures/`, `lessons/`, `vocab/` | `go test ./app/modules/writing/... ./app/modules/library/...` |

A set (`s1` = "Set 1") groups four tests modelled on one book (`book.json` → `source_ref: "cam21"`). A test folder
must hold all 3 reading parts (questions 1–40) and/or all 4 listening parts (1–40) before the whole-tree check passes.

## Workflow

1. **Find the blueprint** (`references/sources.md`): outline the PDF with `scripts/pdf_outline.py`, extract the
   test's pages with `scripts/pdf_pages.py` (text + PNG; scanned books need the PNG), note the answer key page.
2. **Write our version**, one part per file, following the matching reference:
   `references/reading.md`, `references/listening.md`, `references/writing.md`. Start from an existing file of the
   same kind as a template (e.g. `be/seeds/exam/s1/test1/reading-2.json`, `.../listening-1.json`).
3. **Listening only:** build the audio (`python3 scripts/exam-audio/build.py <file>...`), which writes line times,
   cue times, duration and the audio key into the file; then upload (`cd be && go run ./cmd exam-media ../scripts/exam-audio/out/<sha>.mp3`).
4. **Check each file on its own** while the rest of the test is unfinished:
   `cd be && EXAM_PART=exam/<set>/test<N>/<skill>-<P>.json go test ./app/modules/exam/internal/use-cases/ -run TestExamSeedFile -v`
   then the whole tree: `go test ./app/modules/exam/... -count=1` (or `make test`).
5. **Quality pass** with `references/quality.md` (read every answer back against the passage or script).
6. **Load locally:** `make seed` (idempotent; changed files update in place, attempts keep their parts). Try a part
   in the app (`/reading`, `/listening`, `/writing/task-1`).
7. Update the epic plan's status line (`pdm/epic10-reading-listening/…` or `pdm/epic5-writing-practice/…`).

Several tests can be written in parallel by sub-agents (one test folder each); give each the blueprint pages, the
template file, the rules in the reference file and the single-file check command.

## Rules that the checks enforce (and ones they cannot)

- Numbering is continuous: reading 1–13 / 14–26 / 27–40 (or as the blueprint splits them), listening 10 per part.
- Every question has an answer and an English `why`; letters exist among the options; gap answers obey the group's
  word limit (`"Charles Darwin"` breaks ONE WORD ONLY); `evidence.quote` is copied **word for word** from the paragraph
  or transcript line it names (only NOT GIVEN may omit evidence). Unknown JSON fields are rejected.
- Listening: every speaker has a voice; line times increase; cues sit inside the audio and name a real line.
- Not checked, so on you: facts are true (only well-established ones; never invent quotes or study results for real
  people), each question has exactly one defensible answer, distractors are fair, the language is natural IELTS
  register, and nothing is copied from the book.
