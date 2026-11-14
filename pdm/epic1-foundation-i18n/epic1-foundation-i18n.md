# Epic 1: Foundation, Authentication & i18n (vi / en)

| Field | Value |
|---|---|
| Status | Implemented (2026-11-14). Media storage uses AWS S3 (`ap-southeast-1`) instead of MinIO; the health check reports `storage`. Playwright E2E not yet added (the `test-login` endpoint for it exists) |
| Depends on | None |
| Blocks | Epic 2, Epic 3, Epic 4 |
| Skills used | `gin-skills` (backend), `nextjs-skills` (frontend), `taste-skill` (UI, installed when UI work starts) |

## 1. Goal

Set up a runnable full-stack skeleton: a Go/Gin API and a Next.js web app with authentication, Vietnamese and English UI, a design system, and local infrastructure. Every later epic only adds feature modules on top of it.

## 2. Scope

**In scope**
- Backend bootstrap from the `feat/go-gin` template in `be/`.
- Frontend bootstrap with `pnpx create-next-app` in `fe/`.
- Local infrastructure with `docker-compose`: PostgreSQL and MinIO (S3-compatible object storage for images and audio).
- Authentication: **Sign in with Google only** (OAuth 2.0 / OpenID Connect), sessions in httpOnly cookies.
- User profile: display name, UI language, timezone, day-start hour.
- i18n for Vietnamese (default) and English, on both frontend and backend.
- App shell (navigation, layout, light/dark theme) and design tokens.

**Out of scope**
- Email/password accounts, email verification and password reset. Google handles all of these.
- Payments and subscriptions.
- Native mobile apps (the web app must be mobile-first, though).
- Admin panel UI (admin tasks use CLI scripts in v1).

## 3. Architecture overview

```
┌──────────────── fe/ (Next.js App Router) ────────────────┐
│  next-intl ([locale] routing) · shadcn/ui · React Query    │
│  Zustand (UI state) · RHF + Zod · Server Actions (cookies) │
└──────────────────────────┬─────────────────────────────────┘
                           │ REST JSON /api/v1
┌──────────────────────────▼─────────────────────────────────┐
│  be/ (Gin · Uber Fx · GORM · Zerolog)                      │
│  modules: auth, users, (decks, study, vocab, collections)  │
└───────┬───────────────────────────┬────────────────────────┘
        ▼                           ▼
   PostgreSQL                  MinIO / S3 / R2 (media)
```

- Backend modules follow `gin-skills`: `domain/`, `infrastructure/`, `internal/`, `dto/`, `ports/`, `module.go`; SQL migrations in `migrations/`; `AppError` plus `ErrorMiddleware`; rate limiting.
- Frontend follows `nextjs-skills`: `src/features/<feature>/{api,components,hooks,schemas}`, shadcn components first, no hard-coded colours, `T`-prefixed types.

## 4. User stories

### US-1.1 Backend bootstrap
**As a** developer **I want** a backend cloned from the team template **so that** every module follows the same architecture.

Acceptance criteria
- [ ] `be/` cloned from `longtk26/template-frameworks` branch `feat/go-gin`, with fresh git history.
- [ ] `GET /api/v1/health` returns `{status:"ok", db:"ok", storage:"ok"}`.
- [ ] Config is read from `.env`: DB, storage, JWT secrets, CORS origins, Google OAuth (`OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET`, `OAUTH_REDIRECT_URL`), and API keys (`PIXABAY_API_KEY`, `UNSPLASH_ACCESS_KEY`, `UNSPLASH_SECRET_KEY`, `UNSPLASH_APP_ID`, Anthropic, TTS). The variable names match the existing root `.env`.
- [ ] `make mgup` / `make mgdown` work against the compose database.
- [ ] Global rate limit and error middlewares are active. Errors follow the format in §5.
- [ ] CORS allows the frontend origin with credentials.

### US-1.2 Frontend bootstrap
**As a** developer **I want** a Next.js app with the agreed stack **so that** feature work can start immediately.

Acceptance criteria
- [ ] `fe/` created with `pnpx create-next-app@latest` (TypeScript, App Router, Tailwind, `src/`), using pnpm only.
- [ ] shadcn/ui initialised. Semantic tokens live in `globals.css` for light and dark themes.
- [ ] React Query provider, Zustand (with `devtools`), RHF + Zod, and an API client wrapper that sends credentials.
- [ ] Layout primitives exist: `Container`, `Stack`, `Grid`.

### US-1.3 Internationalisation (vi / en)
**As a** Vietnamese learner **I want** the whole interface in Vietnamese **so that** I can use it comfortably. **As an** English speaker, I can switch to English.

Acceptance criteria
- [ ] `next-intl` with locale-prefixed routes: `/vi/...` and `/en/...`. `/` redirects based on the `NEXT_LOCALE` cookie, then the browser's `Accept-Language`, then falls back to `en`.
- [ ] A language switcher in the header keeps the current page. The choice is stored in the `NEXT_LOCALE` cookie and, when logged in, in the user profile.
- [ ] Message catalogues live in `fe/messages/vi.json` and `fe/messages/en.json`, namespaced by feature (`common`, `auth`, `decks`, `study`, ...).
- [ ] No hard-coded user-facing strings in components. A lint or CI script fails when a key exists in one catalogue but not the other.
- [ ] Dates, numbers and relative times ("in 3 days" / "3 ngày nữa") use `Intl` through next-intl formatters.
- [ ] The backend never returns translated prose. It returns a stable **error code** plus parameters, and the frontend translates them (see §5).
- [ ] Page `<title>`, meta description and `hreflang` alternates are localised.
- [ ] **Content language vs UI language:** card content (the Vietnamese meaning) does not change when the UI language changes. Data models still store `meaning_lang` so other meaning languages can be added later.

### US-1.4 Sign in with Google
**As a** learner **I want** to sign in with my Google account in one click **so that** I don't need another password and my progress is saved across devices.

**Flow** (Authorization Code + PKCE; the client secret never leaves the backend):

```
Browser            fe/ (Next.js route handlers)            be/ (Gin)                     Google
  │ click "Continue with Google"
  ├──► GET /api/auth/google/start?next=/vi/dashboard
  │        create state, nonce, PKCE verifier → short-lived httpOnly cookie (10 min)
  │◄── 302 to accounts.google.com (scope: openid email profile, prompt=select_account)
  ├──────────────────────────────────────────────────────────────────────────────────► consent
  │◄─────────────────────────────── 302 /api/auth/google/callback?code&state ◄──────────┤
  ├──► GET /api/auth/google/callback
  │        check state cookie
  │        POST /api/v1/auth/google {code, code_verifier, nonce, redirect_uri} ──►│
  │                                         exchange code (client secret) ───────►│
  │                                         verify ID token: signature (Google JWKS),
  │                                         aud = OAUTH_CLIENT_ID, iss, exp, nonce,
  │                                         email_verified = true
  │                                         upsert user by google_sub
  │        ◄── {access_token, refresh_token, user}
  │        set httpOnly cookies via cookies(), clear OAuth cookie
  │◄── 302 to `next` (relative path only, to prevent open redirects)
```

Acceptance criteria
- [ ] The login page has a single **"Continue with Google"** button (official Google button style), translated in vi and en.
- [ ] Authorised redirect URIs are registered in Google Cloud Console: `http://localhost:3000/api/auth/google/callback` for dev, plus the production URL. `OAUTH_REDIRECT_URL` must match exactly.
- [ ] The backend uses `golang.org/x/oauth2` (Google endpoint) for the code exchange and `github.com/coreos/go-oidc/v3` to verify the ID token. A failed check returns `AUTH_GOOGLE_INVALID_TOKEN`.
- [ ] **First login creates the user** from the Google profile (`sub`, email, name, picture). `ui_locale` is taken from the current URL locale and `timezone` from the browser. Later logins update the name and avatar.
- [ ] Users are identified by Google `sub`, not by email, because an email address can change.
- [ ] Our session: an access token (JWT, 15 min) and a refresh token (opaque, 30 days, stored hashed, rotated on every use), both in **httpOnly, Secure, SameSite=Lax, Path=/** cookies set by Next.js (`cookies()`). Nothing goes in `localStorage`. We do not store Google's tokens; we only need the identity.
- [ ] **Same origin:** Next.js rewrites `/api/v1/*` to the Gin backend, so browser requests carry the cookies and no cross-site cookie setup is needed. The Gin auth middleware reads the access token from the `access_token` cookie, or from `Authorization: Bearer` for server-side calls.
- [ ] Silent refresh: on `401 AUTH_TOKEN_EXPIRED`, the API client calls the refresh Server Action once and retries the request.
- [ ] Log out revokes the refresh token, clears the cookies and redirects to the landing page.
- [ ] Middleware redirects unauthenticated users from protected routes to `/[locale]/login?next=...`.
- [ ] Errors (user cancelled consent, state mismatch, Google unreachable) show a translated message on the login page with a "Try again" button.
- [ ] The auth endpoints are rate-limited per IP.
- [ ] **Testing:** a `POST /api/v1/auth/test-login` endpoint, compiled or enabled **only when `APP_ENV=test`**, lets Playwright sign in without Google.

### US-1.5 User profile & study settings
**As a** learner **I want** my timezone and study preferences saved **so that** "today's" reviews are computed correctly.

Acceptance criteria
- [ ] Profile fields: `display_name`, `ui_locale` (`vi` | `en`), `timezone` (IANA, auto-detected from the browser on first login, editable), `day_start_hour` (0–23, default 4).
- [ ] Study defaults, used by Epic 2: `new_cards_per_day` (20), `max_reviews_per_day` (200), `desired_retention` (0.90), `audio_autoplay` (on), `accent` (`uk` | `us`, default `uk`).
- [ ] Settings page built with RHF + Zod and fully translated.

### US-1.6 App shell & design system
**As a** learner **I want** a clean, fast, mobile-friendly interface **so that** studying daily feels effortless.

Acceptance criteria
- [ ] Before any UI work starts, install the taste skill: `npx skills add https://github.com/Leonxlnx/taste-skill`. Note: `design-taste-frontend` already exists in `~/.claude/skills`; check whether it is the same package before installing twice.
- [ ] Outside-in build: shell, then regions, then leaf components.
- [ ] Navigation: **Dashboard**, **My decks**, **Collections**, **Add words**, **Settings**. Bottom tab bar on mobile, sidebar on desktop.
- [ ] Light and dark themes come from tokens only. No hard-coded hex or arbitrary colour classes.
- [ ] Vietnamese-capable fonts with full diacritic support (for example Be Vietnam Pro or Inter) loaded via `next/font`.
- [ ] A public landing page for logged-out users (short pitch, CTA to register), in both languages.
- [ ] Accessibility: keyboard navigable, visible focus, WCAG AA contrast.

## 5. API conventions (apply to all epics)

- Base path `/api/v1`, JSON, `snake_case` fields, timestamps in UTC ISO-8601.
- Error response:
  ```json
  { "error": { "code": "AUTH_INVALID_CREDENTIALS", "message": "dev-only english text", "params": {}, "fields": { "email": "VALIDATION_EMAIL" } } }
  ```
  The frontend maps `code` and `fields.*` to translation keys `errors.<CODE>`.
- Pagination: `?page=&page_size=` returns `{ items, page, page_size, total }`.
- The language header (`Accept-Language`) is accepted but only used for emails.

## 6. Data model (initial migrations)

```sql
users(id uuid pk, google_sub text unique not null, email citext not null, display_name, avatar_url,
      ui_locale text default 'vi', timezone text default 'Asia/Ho_Chi_Minh', day_start_hour smallint default 4,
      last_login_at, created_at, updated_at)
user_study_settings(user_id pk fk, new_cards_per_day int default 20, max_reviews_per_day int default 200,
      desired_retention numeric(3,2) default 0.90, learning_steps text default '1m 10m',
      relearning_steps text default '10m', audio_autoplay bool default true, accent text default 'uk',
      show_image_on_front bool default false)
refresh_tokens(id, user_id, token_hash, expires_at, revoked_at, replaced_by, user_agent, ip, created_at)
```

## 7. Non-functional requirements

- Lighthouse on mobile for the landing and dashboard pages: Performance ≥ 90, Accessibility ≥ 95.
- All secrets come from env. `.env.example` is committed; `.env` is git-ignored.
- Structured JSON logs with a request ID (Zerolog), propagated to the frontend in the `X-Request-Id` header.

## 8. Open questions for review

1. ~~Default locale~~ **Decided:** detect from the browser (`Accept-Language`). Fall back to `en` when the browser prefers neither vi nor en. A manual choice is remembered in the `NEXT_LOCALE` cookie.
2. ~~Hosting target~~ **Decided:** frontend on **Vercel**, backend and PostgreSQL on **Railway**. Vercel rewrites `/api/v1/*` to the Railway URL, so cookies stay first-party. The production Vercel domain must be added as a Google authorised redirect URI.
3. **Guest mode:** let visitors try 5 cards before signing in? *Proposal: later.*

## 9. Definition of done

- `docker compose up` followed by `make run` (be) and `pnpm dev` (fe) gives a working app where a user can sign in with a real Google account, switch vi/en, edit settings and log out.
- The checklists in `gin-skills` and `nextjs-skills` pass.
- Unit tests for the auth use cases, including ID-token checks (wrong `aud`, expired, bad nonce, unverified email) and refresh-token rotation and reuse detection. One end-to-end smoke test (Playwright, test-login) covering sign in → switch language → log out.
