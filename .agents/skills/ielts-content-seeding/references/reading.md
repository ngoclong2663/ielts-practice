# Reading parts

One file per passage: `be/seeds/exam/<set>/test<N>/reading-<P>.json`. A new set also needs
`be/seeds/exam/<set>/book.json`:

```json
{ "slug": "s2", "source_ref": "cam20", "title": { "vi": "Bộ đề 2", "en": "Set 2" }, "sort": 2 }
```

The slug must equal the folder name; part slugs are derived (`s2-t1-r1`). Templates:
`be/seeds/exam/s1/test1/reading-1.json` (notes + T/F/NG, unlettered), `reading-2.json` (lettered sections,
matching), `reading-3.json` (book review, MCQ + word bank + Y/N/NG), and `s1/test2/reading-1.json` (table).
The Go format is `SeedPart` in `be/app/modules/exam/internal/use-cases/seed-exam.go` and
`be/app/modules/exam/domain/entity/content.go`.

## File shape

```jsonc
{
  "source_ref": "cam20-test1-reading1",       // internal only
  "title": "…", "subtitle": "…",             // subtitle optional
  "topic": "environment",                     // free text, used for labels later
  "q_from": 1, "q_to": 13,
  "paragraphs": [{ "id": "A", "label": "A", "text": "…" }],   // unlettered: { "id": "p1", "text": "…" }
  "groups": [ … ],
  "key": { "1": { "answer": ["…"], "evidence": { "block": "A", "quote": "exact words" },
                  "pairs": [["question words", "passage words"]], "why": "…", "trap": "…" } }
}
```

## Group types

| `type` | Answer | Fields |
|---|---|---|
| `tfng`, `ynng` | button | `items: [{n, text}]`; answers `TRUE`/`FALSE`/`NOT GIVEN` or `YES`/`NO`/`NOT GIVEN` |
| `mcq` | one letter | `items: [{n, text, options: [{key, text}]}]` (≥3 options) |
| `choose_n` | N letters, any order | `stem`, `pick` (= number of questions), `options`; every question's key has the same set |
| `match_info` | letter of a paragraph | `options: [{key:"A"}…]`, `items`, `reuse_options: true` if "any letter more than once" |
| `match_features`, `match_endings`, `match_headings` | letter / roman numeral | `options` with text (people, endings, headings `i`…`x`), `items`, optional `title` ("List of people") |
| `summary_bank` | letter per gap | `options` (word bank A–I with text), `lines` with `{n}` gaps |
| `notes`, `sentence`, `summary`, `flowchart` | typed words | `lines` with `{n}`; `word_limit`, `allow_number`; optional `title` |
| `table` | typed words | `header`, `rows` (cells with `{n}`), `word_limit`, `allow_number` |
| `short_answer` | typed words | `items: [{n, text}]`, `word_limit` |

Section headings: in `notes`, a line starting with `## ` is a heading (`"## Threats"`); in `table` and `form`, a row
with a single cell is a heading row across the table (`["Opinions about the library"]`).

Word limits: "ONE WORD ONLY" → `word_limit: 1`; "ONE WORD AND/OR A NUMBER" → `1` + `allow_number: true`;
"NO MORE THAN TWO WORDS" → `2`; "A NUMBER" → `0` + `allow_number`. Answers in book notation are expanded:
`"30/thirty"`, `"sport(s)"`, `"(the) harbour"`; list UK and US spellings.

## Writing the passage

- 750–950 words, Academic register, vocabulary around band 6–7.5; lettered paragraphs only when a group needs them.
- Order gap answers as they appear in the text (IELTS completion questions follow passage order); T/F/NG and
  matching questions need not.
- Each FALSE must contradict a specific sentence; each NOT GIVEN must touch the topic without settling it (and
  make sure nothing elsewhere settles it); each TRUE paraphrases rather than repeats.
- Explanations (English, plain, for band 5): `why` in 1–3 sentences, `pairs` showing the paraphrase (the main
  teaching point), `trap` when a distractor is close (a word that appears in the text but answers a different
  question, the first number mentioned, an option that is true but not the main point).
- Mix of answers: roughly even T/F/NG, letters not in alphabetical order, distractor options in word banks.
