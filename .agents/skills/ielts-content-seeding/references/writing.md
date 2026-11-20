# Writing tasks

A pack per book: `be/seeds/writing/<pack>/` (packs so far: `cam19`, `cam20`, `cam21`). Background:
`pdm/epic5-writing-practice/epic5-writing-practice.md` §6 (tasks, figures, seeding) and
`pdm/epic8-vocab-library/epic8-vocab-library.md` §4 (seeded lessons and vocabulary). Use `cam21` as the template.

```
be/seeds/writing/<pack>/
  pack.json                       # the tasks (our wording), with Task 1 "facts" (the figure as data)
  figures/test<N>-task1.svg       # our own redrawn figures with changed data
  lessons/<task>.json             # Task 1 lesson; Task 2: <task>.<position>.json (agree, disagree, partly, …)
  vocab/<task>.json               # useful vocabulary, each term a full card or a card_ref to the library
```

## Steps

1. Blueprint each test's Task 1 (figure type, what it shows) and Task 2 (question type, topic) from the book.
2. `pack.json`: rewrite every prompt in our own words. Task 1 keeps the task type with **changed data, labels,
   years or layout**; `facts` holds the figure as data (series, table cells, process stages, map changes).
   Task 2 keeps the topic with our own wording and lists `positions`. `source_ref` is internal only.
3. Figures: `python3 scripts/writing-figures/build.py <pack>` draws charts from `facts` (add `<pack>.py` there for
   maps and processes, as `cam19.py` and `cam20.py` do).
4. Lessons and vocabulary (Epic 8): write one lesson per Task 1 task and per Task 2 position, and one vocab set per
   task, following the cam21 files (plan questions, paragraphs of sentences at every support level, model answer;
   vocab groups `topic` / `paraphrase` / `task_language` with full cards). Each lesson carries the task's
   `task_hash`: run the check below and copy the hash it prints. Pick photos for new cards with
   `cd be && go run ./cmd library-images <vocab file>`.
5. Check: `cd be && go test ./app/modules/writing/... ./app/modules/library/... -count=1`
   (`TestCommittedWritingSeeds` names any wrong `task_hash`, unknown task or card problem), then `make seed`;
   figures go to S3 under content-addressed keys.
6. Review in the app: open each task at level 1 and 4 and read the lesson and the vocabulary step.
