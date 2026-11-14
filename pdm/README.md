# Product Plan: IELTS Practice (Vocabulary first)

## Vision

An IELTS learning website. The first feature is **vocabulary learning with Anki-style spaced repetition**, made richer than Anki with **audio, images and auto-generated card content**. The interface is available in **Vietnamese and English**.

## Epics

| # | Epic | Outcome | Depends on |
|---|---|---|---|
| 1 | [Foundation, Auth & i18n](epic1-foundation-i18n/epic1-foundation-i18n.md) | Runnable Gin + Next.js app, Google sign-in, vi/en, design system | None |
| 2 | [Decks & SRS Study Engine](epic2-srs-study-engine/epic2-srs-study-engine.md) | Anki-faithful FSRS scheduler + study screen with audio/image | 1 |
| 3 | [Smart Card Generation](epic3-smart-card-generation/epic3-smart-card-generation.md) | Paste words → rich cards (forms, vi meaning, collocations, examples, image, audio) | 1, 2 |
| 4 | [Vocabulary Collections](epic4-vocabulary-collections/epic4-vocabulary-collections.md) | Curated IELTS lists; "learn N words today" with a custom title | 2, 3 |

Background research: [docs/features/vocabulary/anki-srs-research.md](../docs/features/vocabulary/anki-srs-research.md)

## Stack

- **Backend** (`be/`): Go, Gin, Uber Fx, GORM, PostgreSQL, Zerolog, raw SQL migrations (`gin-skills`)
- **Frontend** (`fe/`): Next.js App Router, pnpm, shadcn/ui, React Query, Zustand, RHF + Zod, next-intl (`nextjs-skills`, `taste-skill` for UI)
- **Media:** S3-compatible storage (MinIO locally) + CDN
- **External services:** Anthropic Claude (card content), Unsplash / Pixabay (images), cloud TTS (audio)

## Top decisions needed before implementation

1. Epic 3 Q1: meaning of "one example per word" (per word form, or per pasted word).
2. Epic 3 Q3: TTS provider (cloud neural TTS vs browser-only).
3. Epic 4 Q1: source and licence of the "4000 IELTS words" list.
4. Epic 4 Q2: one deck per day (proposed) vs one deck per collection.
5. Epic 2 Q1: reverse cards (vi → en) in v1 or later.
6. Epic 1 Q2: hosting target.

Decided: authentication is **Google sign-in only** (Epic 1, US-1.4).
