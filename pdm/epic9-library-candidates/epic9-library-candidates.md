# Epic 9: Library Candidates

| Field | Value |
|---|---|
| Status | Built (2026-11-19) on `feat/epic9-library-candidates`: admin tab "Thư viện" (`/admin/library`), candidates API, approve/reject (one or many, Shift-click ranges), approved words survive a re-seed |
| Depends on | Epic 3 (generated cards, lexeme cache), Epic 6 (admin board), Epic 8 (library) |

## 1. Goal

Grow the library from what learners actually add. A word pasted in Add words that the library does not have gets an AI card, cached for every learner (Epic 3). When enough different learners keep that card in a deck, it becomes a **candidate**. The owner reviews candidates on the admin board and approves each one into a library topic; an approved card is locked like the seeded ones.

## 2. Scope (lean first version)

**In scope**
- Candidates: AI-made cards (`lexemes.origin = 'generated'`, ready) saved to a deck by at least `LIBRARY_CANDIDATE_MIN_USERS` different learners (default 3), not yet approved or rejected.
- Admin page "Ứng viên thư viện": card preview, number of learners, reports, a topic picker (the 14 Epic 4 topics plus "Khác") already set to the word's topic, Approve / Reject per word, and Approve / Reject for several ticked words at once (each into its own topic, or all into one chosen topic).
- **Topic tagging:** the card generator (Epic 3) also returns `topic`, one of the 14 topics or `other` for general language, in the same call (no extra AI request). It is stored on the lexeme (`lexemes.topic`, migration `20261119000018`) and preselected for the candidate. Cards generated before this get a guess from the words they share with each topic's library words and description (`library/domain/rules/topic.go`). Live test: "sightseeing" → travel, "decide" → other, both models.
- Approving adds the word to the chosen topic of the collection (`source = 'learner'`) and locks its card (`origin = 'library'`). It shows in the library at once.
- A later `api seed` never removes approved words from a topic.

**Out of scope (later)**
- Editing a card before approving it (approve as is, or reject).
- Approving without review after many uses.
- Writing approved words back to the seed files.

## 3. Data

- Migration `20261119000017`: `library_candidate_decisions (lexeme_id PK, decision approved|rejected, topic_id, decided_at)`. A rejected word is not shown again.
- `library_words.source` gains `learner`; a topic that does not exist in the collection yet (e.g. "travel" or "other") is created when its first word is approved.

## 4. API (admin token)

| Method & path | Purpose |
|---|---|
| `GET /api/v1/admin/library/candidates` | Candidates (most learners first), the threshold, the collections and the allowed topics |
| `POST /api/v1/admin/library/candidates/approve` | `{collection, items: [{lexeme_id, topic_id}]}` |
| `POST /api/v1/admin/library/candidates/reject` | `{lexeme_ids: [...]}` |
