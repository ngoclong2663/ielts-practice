# CLAUDE.md

IELTS practice site: vocabulary with Anki-style spaced repetition (FSRS) and AI-generated cards, plus guided writing practice. Vietnamese and English UI. The README holds setup, env variables and deployment.

## Layout

| Folder | What |
|---|---|
| `be/` | Go API: Gin, Uber Fx, GORM, PostgreSQL, raw SQL migrations in `be/migrations/` |
| `fe/` | Next.js 16 App Router, next-intl (`fe/messages/en.json`, `vi.json`), shadcn/ui, React Query |
| `pdm/` | One plan per epic (`pdm/epicN-<name>/epicN-<name>.md`), indexed in `pdm/README.md` |
| `.agents/skills/` | Project conventions: `gin-skills` (backend), `nextjs-skills` (frontend). Read the relevant one before writing code |

Backend modules live in `be/app/modules/<module>/` with `domain/` (entities, rules, repository interfaces), `ports/in`, `ports/out`, `infrastructure/<adapter>/`, `internal/` (router, presenter, one use case per file in `internal/use-cases/`) and `dto/`. New modules are registered in `be/app/container.go`.

Frontend features live in `fe/src/features/<feature>/` (`api/`, `components/`, `hooks/`, `types.ts`, `query-keys.ts`). Next.js 16 differs from older versions: check `fe/node_modules/next/dist/docs/` before using an API (see `fe/AGENTS.md`).

## Commands (from the repo root)

```bash
make be      # migrations + seed, then the API on :8080
make fe      # web app on :3000
make test    # Go unit + integration tests, then fe lint, typecheck and i18n check
make e2e     # Playwright on an isolated API, web and database
```

Every new UI string goes in both `en.json` and `vi.json` (`pnpm i18n:check` fails otherwise). Backend error codes in `be/pkg/exceptions/exception.go` are stable and translated on the frontend as `errors.<CODE>`.

## Git conventions

- Commit messages: Conventional Commits with a scope, e.g. `feat(writing): ...`, `fix(fe): ...`, `refactor(be): ...`. The subject is short and imperative. The body is a bullet list of what changed and why.
- **No watermark in commits**: never add `Co-Authored-By: Claude ...`, "Generated with Claude Code" or any other AI attribution line, either to commit messages or to PR descriptions.
- An epic is developed on its own branch (`feat/epicN-<name>`), never committed straight to `main`.
- Never commit `.env` files. Only the `*.env.example` files are committed.
