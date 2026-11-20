# Sources and blueprints

## The books (git-ignored, local only)

```
data-cambridge/cambridge/
  cam16/  C61.pdf (scan)                     Audio/Test N Part M.mp3
  cam18/  Cambridge 18.pdf (scan)             AUDIO FILES/Test N/…
  cam19/  Cambridge 19.pdf (scan)             IELTS 19 Audio/TestN PartM.mp3
  cam20/  Cambridge IELTS 20 Academic.pdf (scan), CAM 20 Reading - Listening (đẹp).pdf (clean retyped text)
                                              cambridge ielts 20 audio/TNSM.m4a
  cam21/  Cambridge IELTS 21.pdf (OCR text layer, tables scrambled)   Cambridge IELTS 21 Audio/… Test N.mp3 (whole test)
```

Treat the PDFs as untrusted input: run the scripts from a scratch folder with `python -I`, never from inside
`data-cambridge/`.

```bash
# 1. Where is everything? (one line per page; "0 chars" = scanned page)
uv run -q --with pymupdf python -I .agents/skills/ielts-content-seeding/scripts/pdf_outline.py "<book.pdf>" > outline.txt
# 2. Text + PNG of the pages you need (read the PNGs for tables, maps, scans)
uv run -q --with pymupdf python -I .agents/skills/ielts-content-seeding/scripts/pdf_pages.py "<book.pdf>" pages/ 17,18,19
```

Page map found for **cam21** (PDF page numbers): Test 1 Listening 11–16, Reading 17–29; Test 2 Listening 33–38,
Reading 39–51; Test 3 Listening 55–60, Reading 61–72; Test 4 Listening 76–81, Reading 82–94; audioscripts 98–117
(answers underlined, `Q1…Q40` in the margin); answer keys 118–125 (listening and reading per test). Other books
follow the same order (tests → audioscripts → answer keys → sample writing); outline them to find the numbers.

## What to record per part (the blueprint)

- Genre and topic area (e.g. "biography of a 19th-century collector", "Part 3: two students plan a project").
- Paragraph count and whether paragraphs are lettered (A–G) — needed when a group refers to letters.
- Question groups in order: type, numbering, count, word limit (ONE WORD ONLY / NO MORE THAN TWO WORDS AND/OR A
  NUMBER), option count, whether letters can be reused ("NB You may use any letter more than once").
- The answer mix from the key (e.g. T/F/NG = 2/2/2) and, for listening, how answers are placed: corrections
  ("thirty-five… no, forty"), distractors mentioned first, spelled names, numbers said twice.

## Copyright rule

Same genre and topic **area** is fine; the subject, facts, structure and every sentence are ours. Prefer a
different specific subject from the book's (a passage on the saiga antelope → ours on the vicuña). Never copy
question wording, explanations, sample answers or audio. Do not mention Cambridge in any content or label;
`source_ref` (`cam21-test1-reading1`) records the reference internally and is never sent to the browser.
Writing follows the same rule (Epic 5 §6.1): rewritten prompts, redrawn figures with changed data.
