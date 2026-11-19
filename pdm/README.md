# Product Plan: IELTS Practice (Vocabulary first)

## Vision

An IELTS learning website. The first feature is **vocabulary learning with Anki-style spaced repetition**, made richer than Anki with **audio, images and auto-generated card content**. The interface is available in **Vietnamese and English**.

## Epics

| # | Epic | Outcome | Depends on |
|---|---|---|---|
| 1 | [Foundation, Auth & i18n](epic1-foundation-i18n/epic1-foundation-i18n.md) | Runnable Gin + Next.js app, Google sign-in, vi/en, design system | None |
| 2 | [Decks & SRS Study Engine](epic2-srs-study-engine/epic2-srs-study-engine.md) | Anki-faithful FSRS scheduler + study screen with audio/image | 1 |
| 3 | [Smart Card Generation](epic3-smart-card-generation/epic3-smart-card-generation.md) | Paste words → rich cards (forms, vi meaning, collocations, examples, image, read aloud by the browser) | 1, 2 |
| 4 | [Words by Topic](epic4-topic-words/epic4-topic-words.md) | Pick a topic, level and word count; review suggested new words (no duplicates in the deck), then generate cards with the Epic 3 flow. Replaces the dropped "Collections" plan | 2, 3 |
| 5 | [Guided Writing Practice](epic5-writing-practice/epic5-writing-practice.md) | IELTS Writing Task 1 & 2 step by step: plan in Vietnamese, write each paragraph sentence by sentence with feedback, then the full answer with an estimated band. Task bank seeded pack by pack | 1, 3 |
| 6 | [Feedback & Admin Board](epic6-feedback-admin/epic6-feedback-admin.md) | Learners send feedback from the account menu; each one is posted to a Telegram group. Owner-only admin board: usage overview, users with activity and AI use, feedback to read and resolve | 1 (reads 2–5) |
| 7 | [Useful Vocabulary, AI Cost & Guide Videos](epic7-useful-vocab/epic7-useful-vocab.md) | Writing lessons get topic words, paraphrases and task language with synonyms, saved to decks in one step. Admin board shows Claude and Gemini cost per model. Short narrated guide videos (Anki-style learning), stored on S3 | 3, 5, 6 |
| 8 | [Vocabulary Library & Prompt Safety](epic8-vocab-library/epic8-vocab-library.md) | A library of reviewed, ready-made cards (IELTS first: 10 topics × 30 words) added to a deck in one step with no AI call; writing lessons and vocabulary (as full cards) seeded and reviewed once; "type by hand" tab removed; learner text in prompts hardened against injection | 3, 4, 5, 7 |

Background research: [docs/features/vocabulary/anki-srs-research.md](../docs/features/vocabulary/anki-srs-research.md)

## Stack

- **Backend** (`be/`): Go, Gin, Uber Fx, GORM, PostgreSQL, Zerolog, raw SQL migrations (`gin-skills`)
- **Frontend** (`fe/`): Next.js App Router, pnpm, shadcn/ui, React Query, Zustand, RHF + Zod, next-intl (`nextjs-skills`, `taste-skill` for UI)
- **Media:** card images hotlinked from Unsplash / Pixabay; S3 holds writing task figures and future user uploads
- **External services:** Anthropic Claude (card content), Unsplash / Pixabay (images); audio is the browser's Web Speech API

## Top decisions needed before implementation

5. Epic 2 Q1: reverse cards (vi → en) in v1 or later.
6. Epic 1 Q2: hosting target.

Decided: authentication is **Google sign-in only** (Epic 1, US-1.4). Epic 3: one example per word form (proposal), browser speech only (no TTS), provider images hotlinked (no S3).
