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

Background research: [docs/features/vocabulary/anki-srs-research.md](../docs/features/vocabulary/anki-srs-research.md)

## Stack

- **Backend** (`be/`): Go, Gin, Uber Fx, GORM, PostgreSQL, Zerolog, raw SQL migrations (`gin-skills`)
- **Frontend** (`fe/`): Next.js App Router, pnpm, shadcn/ui, React Query, Zustand, RHF + Zod, next-intl (`nextjs-skills`, `taste-skill` for UI)
- **Media:** images hotlinked from Unsplash / Pixabay; S3 reserved for future user uploads
- **External services:** Anthropic Claude (card content), Unsplash / Pixabay (images); audio is the browser's Web Speech API

## Top decisions needed before implementation

5. Epic 2 Q1: reverse cards (vi → en) in v1 or later.
6. Epic 1 Q2: hosting target.

Decided: authentication is **Google sign-in only** (Epic 1, US-1.4). Epic 3: one example per word form (proposal), browser speech only (no TTS), provider images hotlinked (no S3).
