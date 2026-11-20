# Quality pass (do this for every part before committing)

The automated checks prove the files are well formed. These points they cannot prove:

- [ ] **Original.** No sentence, question or explanation is taken from the book; the subject differs from the
      book's; "Cambridge" appears nowhere in content.
- [ ] **True.** Facts are well established; uncertain details are phrased loosely or removed; no invented quote,
      survey figure or study result is attributed to a real person or organisation.
- [ ] **One answer.** Read each question against the text: exactly one option or word fits, and the gap answer
      fits the sentence grammatically (plural, article, verb form).
- [ ] **NOT GIVEN really is.** Nothing anywhere in the passage settles it either way.
- [ ] **Fair traps.** Distractors are tempting but clearly wrong once understood; none relies on a typo or trick.
- [ ] **Explanations teach.** `why` says *why* (not just where); `pairs` show the paraphrase; a `trap` names the
      tempting wrong answer when there is one. English a band-5 learner can follow.
- [ ] **Order and spread.** Completion gaps follow the text order; T/F/NG and letters are well mixed.
- [ ] **Listening sounds right.** Listen to the built MP3 of at least one part per test: names and numbers are
      read as intended (fix with `say`), pauses feel natural, speakers are easy to tell apart.
- [ ] **Seen in the app.** Open the part in the player and the review (evidence highlight, transcript, figure).

Automated scan of a whole set (word counts, types, answer mix, gap answers missing from the text, explanations
without pairs, any mention of the book):

```bash
python3 .agents/skills/ielts-content-seeding/scripts/scan_content.py be/seeds/exam/<set>
```
