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
make be     # applies migrations and seed data, runs the API on http://localhost:8080
make fe     # runs the web app on http://localhost:3000
```

`make help` lists the other commands (`install`, `migrate`, `seed`, `test`, `build`).

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
by default), `PIXABAY_API_KEY` and `UNSPLASH_ACCESS_KEY` in `be/.env`. Words Claude cannot answer
are retried once with Gemini; with only one key set, that provider is used alone. Without either key the "Paste a list"
page shows a notice and every word fails with `VOCAB_GENERATOR_NOT_CONFIGURED`.
To try the whole flow offline, set `VOCAB_FAKE_PROVIDERS=true` (canned content and placeholder pictures; ignored in production).
Pictures are hotlinked from Unsplash and Pixabay, never copied to S3. Pixabay links expire after about 24 h, so cards
point at `/api/v1/images/pixabay/<id>`, which redirects to a fresh link and finds a replacement if the photo was removed.
Other settings: `VOCAB_LLM_MODEL`, `VOCAB_LLM_EFFORT`, `VOCAB_WORKERS`, `VOCAB_DAILY_QUOTA`, `VOCAB_DAILY_SUGGESTIONS`
(word lists by topic per user per day, default 5), `VOCAB_IMAGE_PROVIDERS` (see `be/.env.example`).

### Writing task bank (Epic 5)

Writing tasks are seeded from `be/seeds/writing/<pack>/pack.json` plus the figures next to it. The seeds are compiled
into the API binary and applied by `api seed` (`make seed` locally; `make be` runs it too). In production the container
entrypoint runs it after the migrations, so deploying a new pack seeds it (`RUN_SEEDS=false` turns this off).
The seed is idempotent: unchanged tasks are skipped, changed tasks are updated in place, and figures are uploaded to S3
under content-addressed keys (`writing/figures/<sha256>.svg`). Packs are our own rewritten tasks with redrawn figures;
the practice books in `data-cambridge/` (git-ignored) are only a source of ideas and are never copied. See
[Epic 5 §6](pdm/epic5-writing-practice/epic5-writing-practice.md).
Without S3 (local development, e2e) the figure endpoint serves the copy compiled into the binary.
To add a book: write `be/seeds/writing/<camNN>/pack.json` (our own wording and data), draw its Task 1 figures with
`python3 scripts/writing-figures/build.py <camNN>` (add a `<camNN>.py` module there for maps and processes), then run
`make seed`. Packs so far: cam19, cam20, cam21.

Guide videos (Epic 7) are short narrated MP4s shown on the decks page and in writing. Their sources are local only;
an approved video and its poster go to S3 with `go run ./cmd upload-media <video.mp4> <poster.jpg>` (from `be/`,
content-addressed keys `guides/videos/<sha256>.mp4`), and the printed names go into `be/seeds/guides/videos.json`
with the titles and descriptions; `make seed` (and every deploy) loads them. The API serves the files through a
redirect to a short-lived S3 link, like writing figures.

Vocabulary library (Epic 8): reviewed, ready-made cards in `be/seeds/library/<collection>/` (`collection.json` and
one `<topic>.json` per topic). Learners add them to a deck with no AI call. The writing packs also carry reviewed
lessons (`<pack>/lessons/`) and useful vocabulary with a card per term (`<pack>/vocab/`). After editing a card file,
pick photos with `go run ./cmd library-images <file>...` (from `be/`; Unsplash, then Pixabay; no AI), then check the
files with `go test ./app/modules/library/... ./app/modules/writing/internal/use-cases/` and run `make seed`.

Guided writing practice uses the same models as card generation (`VOCAB_LLM_MODEL`, then `VOCAB_FALLBACK_MODEL`):
one shared lesson per task and position, then a check per paragraph and a review of the full answer. Daily limits
per user: `WRITING_DAILY_CHECKS` (default 30) and `WRITING_DAILY_SUBMISSIONS` (default 5). With
`VOCAB_FAKE_PROVIDERS=true` a fake coach answers offline.

### Feedback and admin board (Epic 6)

Learners send feedback from the account menu. Each one is saved and posted to a Telegram group:

1. Create a bot with [@BotFather](https://t.me/BotFather) (`/newbot`) and copy its token.
2. Add the bot to your group and send any message there.
3. Open `https://api.telegram.org/bot<token>/getUpdates` and copy `chat.id` (group ids start with `-100`).
4. Set `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` in `be/.env` (`TELEGRAM_THREAD_ID` for a topic in a forum group).

Without them feedback is still saved and the board shows "Not sent to Telegram". Telegram is never called with
`APP_ENV=test` or `VOCAB_FAKE_PROVIDERS=true`. Limit: `FEEDBACK_DAILY_LIMIT` per user per 24 h (default 10).

The admin board is at `/en/admin` (overview, users, feedback). The admin account is seeded in the database:
set `ADMIN_USERNAME` and `ADMIN_PASSWORD` in `be/.env` (at least 12 characters in production) and run `make seed`
(`make be` and the production entrypoint run it too). The password is stored as a bcrypt hash; change the variable and
seed again to rotate it. Sign-in returns an admin token kept in the httpOnly `admin_token` cookie (`ADMIN_SESSION_TTL`,
default 12 h). It is signed apart from learner sessions, so a learner token never opens the board. Days on the board
are counted in `ADMIN_TIMEZONE` (default `Asia/Ho_Chi_Minh`). See [Epic 6](pdm/epic6-feedback-admin/epic6-feedback-admin.md).

### Media storage (S3)

S3 holds the writing task figures (above) and is reserved for future user uploads. Card pictures are hotlinked, not stored.
Set `S3_BUCKET_NAME`, `AWS_REGION`, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` in `be/.env`. For a local S3-compatible server (MinIO, RustFS) also set `S3_ENDPOINT` and `S3_FORCE_PATH_STYLE=true`. `GET /api/v1/health` reports `storage: ok | down | not_configured`.

## Project layout (backend modules)

Each module in `be/app/modules/<module>/` has:

- `domain/`: `entity/` (entities and value objects), `rules/` (pure business rules: FSRS scheduler, daily queue, normalising, band rounding), and the repository interfaces at the root
- `ports/in/`: what the module offers to other modules (implemented by its use cases)
- `ports/out/`: external services it needs (implemented in `infrastructure/`)
- `infrastructure/<adapter>/`: one sub-package per technology (`persistence/` for GORM, `google/`, `images/`, `coach/`, …)
- `internal/`: presenter, router, mapper and **one use case per file** in `internal/use-cases/` (no service layer)
- `dto/`

See `.agents/skills/gin-skills` (`references/architecture-and-structure.md` §7–9).

| Module | Responsibility |
|---|---|
| `auth` | Google sign-in, sessions (JWT + rotating refresh tokens) |
| `users` | Profile and study settings (`/me`) |
| `decks` | Decks, cards, card browser, bulk actions |
| `study` | FSRS scheduler, daily queue, answer/undo, counts, stats |
| `vocab` | Paste → generated cards: normalising, jobs + background worker, card content and topic words (through `llm`), image search and link refresh, lexeme cache, reports. Words by topic: suggestions, known words, daily limit |
| `writing` | Writing task bank and its seed (`api seed`), guided attempts, lessons, paragraph checks and full-answer review (Epic 5) |
| `guides` | Guide videos (Epic 7): seed file, list per topic, media links through S3 |
| `library` | Vocabulary library (Epic 8): seeded collections and topics, search, adding reviewed cards to a deck; stores the cards of seeded writing vocabulary |
| `feedback` | Learner feedback, its Telegram notification, and the admin feedback list (Epic 6) |
| `admin` | Admin account (seed, login, admin token) and the board numbers: overview and users (read-only) |
| `notify` | Shared: owner notifications (Telegram; no-op when not configured) |
| `storage` | Shared: S3 storage (writing figures, guide videos; future uploads) |
| `llm` | Shared: language models for every module (Claude first, Gemini fallback; `llm/ports/in.IJSONModel`); records every call in the `ai_usage` ledger |
| `health` | `/api/v1/health` |

## Tests and checks

Backend tests live in three places:

| Where | What | Runs |
|---|---|---|
| next to the code (`app/modules/*/domain`, `pkg/…`) | Unit tests | always |
| [`be/tests/integration/`](be/tests/integration/) | API integration tests: embedded PostgreSQL (downloaded on first run), the real HTTP stack and modules, fake Google, fake AI and image providers. One file per feature (`study_test.go`, `vocab_test.go`, `writing_practice_test.go`, …); shared setup in `harness_test.go` | `make test` (skipped by `make test-unit`) |
| [`be/tests/live/`](be/tests/live/) | Opt-in calls to the real external services, keys from `be/.env` | only with the variables below |

```bash
cd be && make test          # unit + integration tests
cd be && make test-unit     # unit tests only
# Real S3 round trip (put, presigned get, delete) against the bucket in be/.env:
cd be && set -a && . ./.env && set +a && S3_INTEGRATION=1 go test ./tests/live/ -run S3
# Real Pixabay and Unsplash calls:
cd be && PROVIDERS_INTEGRATION=1 go test ./tests/live/ -run ImageProvidersLive
# Real Claude + Gemini: card generation, topic suggestions, writing lessons/checks/reviews (costs a few cents):
cd be && LLM_INTEGRATION=1 go test ./tests/live/ -run 'ContentGeneratorsLive|WordSuggesterLive|CoachLive' -v
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
2. Variables: `DATABASE_URL` (reference the Postgres service), `S3_BUCKET_NAME`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `APP_ENV=production`, `JWT_SECRET` (≥ 32 random chars), `OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET`, `OAUTH_REDIRECT_URL=https://<vercel-domain>/api/auth/google/callback`, `FRONTEND_URL=https://<vercel-domain>`, `TRUSTED_PROXIES=*`, `ANTHROPIC_API_KEY`, `GEMINI_API_KEY`, `PIXABAY_API_KEY`, `UNSPLASH_ACCESS_KEY`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`, `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`.
3. Migrations run automatically at container start (`RUN_MIGRATIONS=false` disables it). Health check: `/api/v1/health`.

**Web on Vercel**
1. Import the repo with root directory `fe/`.
2. Variables: `API_URL=https://<railway-domain>`, `OAUTH_CLIENT_ID`, `OAUTH_REDIRECT_URL=https://<vercel-domain>/api/auth/google/callback`.
3. Browser calls to `/api/v1/*` are rewritten to the Railway API, so auth cookies stay first-party on the Vercel domain.
