# Listening parts

One file per part: `be/seeds/exam/<set>/test<N>/listening-<P>.json`. Template:
`be/seeds/exam/s1/test1/listening-1.json` (table + notes, cues, voices). Question groups and the key work as in
`reading.md`, plus `form` (rows of `[label, value with {n}]`) and `map` (letter labelling on a figure). The
transcript (`lines`) is hidden from the player and shown, with the answer words underlined, in the review.
Forms and notes can have section headings (a one-cell row, or a `## ` line), as in `reading.md`.

## File shape

```jsonc
{
  "source_ref": "cam20-test1-listening1",
  "title": "Courses at a climbing centre", "topic": "sport",
  "q_from": 1, "q_to": 10,
  "voices": { "NARRATOR": "kokoro:bm_daniel", "SOPHIE": "kokoro:bf_emma", "MAN": "kokoro:bm_george" },
  "lines": [
    { "id": "l1", "speaker": "NARRATOR", "text": "Part 1. You will hear … First, you have some time to look at questions 1 to 6.", "at": 0 },
    { "id": "l2", "speaker": "SOPHIE", "text": "It's Helen Marsh, M-A-R-S-H.", "say": "It's Helen Marsh. M, A, R, S, H.", "at": 0 }
  ],
  "audio": { "key": "", "duration": 0,
             "cues": [{ "at": 0, "seconds": 30, "label": "Questions 1–6", "after": "l1" }] },
  "groups": [ … ], "key": { "1": { "answer": ["…"], "evidence": { "block": "l6", "quote": "exact words" }, … } }
}
```

- `at`, `audio.key`, `audio.duration` and each cue's `at` are filled by the build; write zeros.
- `say` is the spoken form when it must differ from the transcript (spelling, digits of a phone number).
- Cues are exam-mode reading pauses; `after` names the narrator line before the pause. Give one before the first
  group and one before each later group the book gives time for ("Before you hear the rest…").
- End with `NARRATOR: "That is the end of Part N."`. Split monologues into lines of 1–3 sentences.

## Script

Part 1: everyday conversation (form, notes, table). Part 2: one speaker in a social setting (MCQ, map/plan,
matching). Part 3: 2–3 people in an academic discussion (MCQ, choose two, matching). Part 4: a lecture
(notes/summary/flow chart). 700–950 words per part (4–6 minutes). Build in the IELTS features: corrections ("…
thirty-five, sorry, forty-five"), the first number or day heard is often not the answer, options all mentioned
but only one meeting the condition, paraphrase between questions and script. Answers must be words spoken.

## Voices (free, local TTS)

`"voices"` maps every speaker to an engine voice: narrator always `kokoro:bm_daniel`. British `kokoro:bf_emma`,
`kokoro:bf_isabella`, `kokoro:bm_george`, `kokoro:bm_lewis`, `kokoro:bm_fable`; American `kokoro:af_heart`,
`kokoro:af_bella`, `kokoro:am_michael`, `kokoro:am_fenrir`; Australian `melo:EN-AU` (one voice; at most one speaker
per part). Append `@0.95` to slow a voice. Kokoro-82M is Apache 2.0, MeloTTS is MIT; both run on the Mac.

## Map and plan figures

`"figure": "figures/<name>.svg"` on a `map` group (options A–I with empty text, items = places). Draw the SVG by
hand in `be/seeds/exam/<set>/figures/`: `viewBox="0 0 640 440"`, white background, `font-family="system-ui,
sans-serif"`, strokes `#334155`, light fills (`#f1f5f9`, `#e2e8f0`, water `#dbeafe`, green `#dcfce7`), fixed
labelled features that are not answers, letters A–I in small circles. Under 50 KB, no external resources. Figures
are compiled into the API and served at `/api/v1/exam/figures/<sha256>.svg`.

## Audio build and upload

```bash
bash scripts/exam-audio/setup.sh          # once: venvs for Kokoro and MeloTTS (needs brew ffmpeg espeak-ng, uv)
python3 scripts/exam-audio/build.py be/seeds/exam/<set>/test<N>/listening-<P>.json   # ~30 s per part
cd be && go run ./cmd exam-media ../scripts/exam-audio/out/<sha256>.mp3              # to S3 (content-addressed)
```

The build renders each line with its voice (cached by text, so edits only re-render changed lines), joins them
with natural pauses into a 64 kbps mono MP3, and rewrites the seed file with line times, cue times, duration and
`audio.key`. Rebuild after any text change, then upload the new MP3. Listen to at least one part per test.
