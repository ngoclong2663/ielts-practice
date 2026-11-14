# IELTS Practice

Vocabulary learning for IELTS with Anki-style spaced repetition, audio and images. Vietnamese and English UI.

| Part | Stack | Folder |
|---|---|---|
| API | Go, Gin, Uber Fx, GORM, PostgreSQL | [`be/`](be/) |
| Web | Next.js 16 (App Router, Cache Components), next-intl, shadcn/ui, React Query | [`fe/`](fe/) |
| Product plan | Epics and research | [`pdm/`](pdm/), [`docs/`](docs/) |

## Local development

Quick start from the project root (after the one-time setup below), in two terminals:

```bash
make be     # applies migrations, runs the API on http://localhost:8080
make fe     # runs the web app on http://localhost:3000
```

`make help` lists the other commands (`install`, `migrate`, `test`, `build`).

Prerequisites: Go 1.26 (the `go` command downloads it automatically), Node 22 with pnpm, Docker, and the [`migrate`](https://github.com/golang-migrate/migrate) CLI.

```bash
# 1. Database (skip if a Postgres already listens on 5432; then just create the
#    `ielts_practice` database and point DATABASE_URL at it)
docker compose up -d

# 2. API (http://localhost:8080)
cd be
cp .env.example .env        # then fill OAUTH_CLIENT_ID / OAUTH_CLIENT_SECRET
make mgup                   # apply migrations
make run

# 3. Web (http://localhost:3000)
cd fe
cp .env.example .env        # OAUTH_CLIENT_ID, API_URL
pnpm install
pnpm dev
```

### Google sign-in setup

In Google Cloud Console → APIs & Services → Credentials, open the OAuth client (type "Web application") and add the **Authorized redirect URIs**:

- `http://localhost:3000/api/auth/google/callback`
- `https://<your-vercel-domain>/api/auth/google/callback`

`OAUTH_REDIRECT_URL` must match one of them exactly, in both `be/.env` and `fe/.env`.

### Card generation (Epic 3)

Set `ANTHROPIC_API_KEY` (Claude writes the card content), `GEMINI_API_KEY` (fallback model, `gemini-3.1-flash-lite`
by default), `PIXABAY_API_KEY` and `UNSPLASH_ACCESS_KEY` in `be/.env` or the project-root `.env`. Words Claude cannot answer
are retried once with Gemini; with only one key set, that provider is used alone. Without either key the "Paste a list"
page shows a notice and every word fails with `VOCAB_GENERATOR_NOT_CONFIGURED`.
To try the whole flow offline, set `VOCAB_FAKE_PROVIDERS=true` (canned content and placeholder pictures; ignored in production).
Pictures are hotlinked from Unsplash and Pixabay, never copied to S3. Pixabay links expire after about 24 h, so cards
point at `/api/v1/images/pixabay/<id>`, which redirects to a fresh link and finds a replacement if the photo was removed.
Other settings: `VOCAB_LLM_MODEL`, `VOCAB_LLM_EFFORT`, `VOCAB_WORKERS`, `VOCAB_DAILY_QUOTA`, `VOCAB_IMAGE_PROVIDERS` (see `be/.env.example`).

### Media storage (S3)

S3 is reserved for future user uploads; nothing is stored there yet.
Set `S3_BUCKET_NAME`, `AWS_REGION`, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` in `be/.env`. For a local S3-compatible server (MinIO, RustFS) also set `S3_ENDPOINT` and `S3_FORCE_PATH_STYLE=true`. `GET /api/v1/health` reports `storage: ok | down | not_configured`.

## Project layout (backend modules)

Each module in `be/app/modules/<module>/` has `domain/`, `infrastructure/`, `dto/`, `ports/` and `internal/`
with **one use case per file** in `internal/use-cases/` (no service layer). See `.agents/skills/gin-skills`.

| Module | Responsibility |
|---|---|
| `auth` | Google sign-in, sessions (JWT + rotating refresh tokens) |
| `users` | Profile and study settings (`/me`) |
| `decks` | Decks, cards, card browser, bulk actions |
| `study` | FSRS scheduler, daily queue, answer/undo, counts, stats |
| `vocab` | Paste → generated cards: normalising, jobs + background worker, Claude content, image search and link refresh, lexeme cache, reports |
| `storage` | S3 storage (reserved for future uploads) |
| `health` | `/api/v1/health` |

## Tests and checks

```bash
cd be && make test          # unit + integration tests (embedded PostgreSQL, downloaded on first run)
cd be && make test-unit     # unit tests only
# Real S3 round trip (put, presigned get, delete) against the bucket in be/.env:
cd be && set -a && . ./.env && set +a && S3_INTEGRATION=1 go test ./app/modules/storage/...
# Real Pixabay and Unsplash calls (keys from be/.env):
cd be && PROVIDERS_INTEGRATION=1 go test ./app/modules/vocab/infrastructure/ -run ImageProvidersLive
# Real Claude + Gemini card generation for 4 words (costs a few cents):
cd be && LLM_INTEGRATION=1 go test ./app/modules/vocab/infrastructure/ -run ContentGeneratorsLive -v
cd fe && pnpm lint && pnpm typecheck && pnpm i18n:check && pnpm build
```

### End-to-end tests

```bash
make e2e
```

Starts a test-mode API on `:8081` (enables `/api/v1/auth/test-login`, so no Google account is needed),
with `VOCAB_FAKE_PROVIDERS=true` (no Claude or image API calls), builds the web app into `fe/.next-e2e` and serves it on `:3001`, then runs Playwright (desktop + mobile)
against a separate `<your database>_e2e` database. It never touches the servers started by `make be` / `make fe`.
First run only: `cd fe && pnpm exec playwright install chromium`. Screenshots: `E2E_SCREENSHOTS=/tmp/shots make e2e`.

## Deployment

**API on Railway**
1. Create a project with a PostgreSQL service and a service from this repo with root directory `be/` (it uses `be/Dockerfile` and `be/railway.json`).
2. Variables: `DATABASE_URL` (reference the Postgres service), `S3_BUCKET_NAME`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `APP_ENV=production`, `JWT_SECRET` (≥ 32 random chars), `OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET`, `OAUTH_REDIRECT_URL=https://<vercel-domain>/api/auth/google/callback`, `FRONTEND_URL=https://<vercel-domain>`, `TRUSTED_PROXIES=*`, `ANTHROPIC_API_KEY`, `GEMINI_API_KEY`, `PIXABAY_API_KEY`, `UNSPLASH_ACCESS_KEY`.
3. Migrations run automatically at container start (`RUN_MIGRATIONS=false` disables it). Health check: `/api/v1/health`.

**Web on Vercel**
1. Import the repo with root directory `fe/`.
2. Variables: `API_URL=https://<railway-domain>`, `OAUTH_CLIENT_ID`, `OAUTH_REDIRECT_URL=https://<vercel-domain>/api/auth/google/callback`.
3. Browser calls to `/api/v1/*` are rewritten to the Railway API, so auth cookies stay first-party on the Vercel domain.
