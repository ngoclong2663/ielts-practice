# Epic 6: Feedback & Admin Board

| Field | Value |
|---|---|
| Status | In progress (2026-11-16) on branch `feat/epic6-feedback-admin` |
| Depends on | Epic 1 (auth, shell, i18n); reads the data of Epics 2–5 |

## 1. Goal

1. Learners can **send feedback** from any page (bug, idea, wrong content, other) in a few seconds.
2. Each new feedback is **posted to a Telegram group** at once, so the owner sees it without opening the app.
3. The owner has an **admin board** (owner only) to see how the app is used (users, decks, cards, reviews, AI use, writing practice, content reports) and to read and resolve feedback.

## 2. Scope

**In scope**
- "Send feedback" dialog in the account menu (desktop and mobile).
- Telegram notification per feedback (Bot API `sendMessage`), off when not configured.
- One admin account seeded in the database (username and bcrypt password hash), signed in through the API.
- Admin pages: Overview, Users, Feedback (status: new → seen → resolved, private note).

**Out of scope (later)**
- Several admin accounts, roles, audit log.
- Actions on users (block, delete, reset limits), editing user content.
- Screenshots or files attached to feedback; replying to the learner in the app.
- Telegram for other events (new sign-ups, content reports, AI errors), cost in money.

## 3. Feedback (learner side)

```
Account menu ▸ Send feedback
  ┌──────────────────────────────────────────┐
  │ Send feedback                             │
  │ (●) Something is broken   ( ) An idea     │
  │ ( ) Wrong word content    ( ) Other       │
  │ ┌──────────────────────────────────────┐ │
  │ │ What happened? What did you expect?   │ │
  │ └──────────────────────────────────────┘ │
  │ We attach the page you are on.  [Send]    │
  └──────────────────────────────────────────┘
```

- Categories: `bug`, `idea`, `content`, `other`. Message: 10–2000 characters.
- The page path and UI language are attached automatically; the API stores the browser's user agent.
- Limit: `FEEDBACK_DAILY_LIMIT` (default 10) per user per 24 h → `FEEDBACK_LIMIT` (429).
- After sending: toast "Thanks, we read every message".

## 4. Telegram notification

- Config (`be/.env`): `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` (group id, e.g. `-1001234567890`), optional `TELEGRAM_THREAD_ID` (a topic in a forum group). Both required to send; otherwise notifications are off (logged once at start).
- Sent in the background after the feedback is saved; a Telegram failure never fails the request. Success sets `feedback.notified_at`, so the board can show "not delivered".
- Never sent in `APP_ENV=test` or with `VOCAB_FAKE_PROVIDERS=true`.
- Message (HTML parse mode, user text escaped, message cut to 1500 characters):

```
🐞 New feedback · Something is broken
From: Learner One (learner@example.com)
Page: /vi/decks/…  ·  vi

The audio button does not play on iPhone…

Open: https://<FRONTEND_URL>/en/admin/feedback?id=<id>
```

Setup: create a bot with @BotFather, add it to the group, send one message in the group, then read the chat id from `https://api.telegram.org/bot<token>/getUpdates`.

## 5. Admin access

The admin account is a row in the database (`admin_accounts`), not a learner account and not a frontend setting.

- **Seed:** `api seed` (run by `make be`, `make seed` and the production entrypoint) upserts one admin from `ADMIN_USERNAME` and `ADMIN_PASSWORD` in `be/.env`. The password is stored as a bcrypt hash (cost 12). Re-running with the same password changes nothing; a new password replaces the hash. With either variable unset the seed skips the admin. Production needs a password of at least 12 characters.
- **Sign-in:** `/[locale]/admin/login` → a Server Action calls `POST /api/v1/admin/auth/login` and stores the returned admin token in the httpOnly cookie `admin_token` (`SameSite=Strict`, 12 h, `ADMIN_SESSION_TTL`). The API rate-limits the endpoint and keeps the timing the same for unknown usernames.
- **Admin token:** an HS256 JWT signed with a key derived from `JWT_SECRET` that differs from the learner key, so a learner token is never accepted on `/api/v1/admin/*` and an admin token is never a learner session. Deleting the `admin_accounts` row revokes access at once.
- **Calls:** the board calls `/api/v1/admin/*` through the same-origin rewrite, so the cookie goes with each request and `AdminAuthMiddleware` checks it. A `401` sends the board back to the sign-in page.
- `proxy.ts` redirects `/admin/*` without the cookie to `/admin/login` (optimistic, like the learner pages). The sign-in page never redirects away, so an expired cookie cannot loop. Admin pages are `noindex`.

## 6. Admin board

Routes: `/admin` (overview), `/admin/users`, `/admin/feedback`. They have their own layout (no learner navigation) with tabs and "Sign out". Days are counted in `ADMIN_TIMEZONE` (default `Asia/Ho_Chi_Minh`).

**Overview**
- Users: total, new in 7 / 30 days, active today / 7 / 30 days. Active means the user studied a card, wrote, generated cards, asked for a word list or signed in.
- Vocabulary: decks, cards (by state: new, learning, review, relearning), reviews (total and 7 days), words in the shared cache.
- AI use: words generated (charged) / served from the cache / failed in 7 days, word lists by topic, writing lessons, paragraph checks, full-answer reviews, tokens in and out in 30 days.
- Writing: attempts, submitted, average estimated band in 30 days.
- Feedback: new / total. Content reports: total, 7 days, words waiting for review.
- A 30-day chart: sign-ups, active users, reviews, AI requests per day.
- The five latest content reports (word, reason, note, who).

**Users**: search by email or name, sort by last active / newest / reviews / AI use; 25 per page. Columns: user, joined, last active, decks, cards, reviews (total, 7 days), AI words, word lists, writing attempts (submitted), writing AI calls, feedback.

**Feedback**: filter by status (new by default) and category. Each item: category, message, page, language, browser, time, who (name, email), Telegram delivered or not. Actions: mark seen, resolve, reopen, private note (≤ 1000 characters). Opening `?id=<id>` (the Telegram link) scrolls to and highlights that item.

## 7. Data model

```sql
CREATE TABLE feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category    VARCHAR(16)  NOT NULL CHECK (category IN ('bug', 'idea', 'content', 'other')),
    message     TEXT         NOT NULL CHECK (char_length(message) BETWEEN 1 AND 2000),
    page        VARCHAR(300) NOT NULL DEFAULT '',
    ui_locale   VARCHAR(8)   NOT NULL DEFAULT '',
    user_agent  VARCHAR(400) NOT NULL DEFAULT '',
    status      VARCHAR(16)  NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'seen', 'resolved')),
    admin_note  VARCHAR(1000) NOT NULL DEFAULT '',
    notified_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_feedback_status ON feedback (status, created_at DESC);
CREATE INDEX idx_feedback_user ON feedback (user_id, created_at);
```

The board reads everything else from the existing tables (`users`, `decks`, `cards`, `review_logs`, `generation_jobs`, `generation_job_items`, `lexemes`, `lexeme_reports`, `topic_suggestion_requests`, `writing_*`).

## 8. Backend

Modules:
- `notify` (shared): `notify/ports/in.IAdminNotifier` with `Notify(ctx, entity.Notification) error`; adapters `infrastructure/telegram` and a no-op one.
- `feedback`: learner endpoint and the admin feedback endpoints.
- `admin`: overview and users (read-only reporting queries).
- `admin` also holds the admin account: seed, login, the admin token and `AdminAuthMiddleware` (`401 ADMIN_UNAUTHORIZED` when the token is missing, invalid, expired or its account is gone).

### 8.1 API

| Method | Path | Auth | Body / query → response |
|---|---|---|---|
| POST | `/api/v1/feedback` | learner | `{category, message, page?, ui_locale?}` → `201 {id, created_at}` |
| POST | `/api/v1/admin/auth/login` | none (rate-limited) | `{username, password}` → `{access_token, expires_at, admin: {id, username}}`; wrong → `401 ADMIN_INVALID_CREDENTIALS` |
| GET | `/api/v1/admin/auth/me` | admin | → `{id, username}` |
| GET | `/api/v1/admin/overview` | admin | → `TAdminOverview` |
| GET | `/api/v1/admin/users` | admin | `?q&sort=last_active\|created\|reviews\|ai&page=1&page_size=25` → `{items, total, page, page_size}` |
| GET | `/api/v1/admin/feedback` | admin | `?status=new\|seen\|resolved\|all&category=&page&page_size` → `{items, total, page, page_size, counts: {new, seen, resolved}}` |
| PATCH | `/api/v1/admin/feedback/:id` | admin | `{status?, admin_note?}` → feedback item |

```jsonc
// GET /api/v1/admin/overview
{
  "generated_at": "2026-11-16T03:00:00Z",
  "timezone": "Asia/Ho_Chi_Minh",
  "users": { "total": 0, "new_7d": 0, "new_30d": 0, "active_1d": 0, "active_7d": 0, "active_30d": 0 },
  "vocab": {
    "decks": 0, "cards": 0,
    "cards_by_state": { "new": 0, "learning": 0, "review": 0, "relearning": 0 },
    "reviews_total": 0, "reviews_7d": 0, "lexemes_cached": 0
  },
  "ai": {
    "words_generated_total": 0, "words_generated_7d": 0, "words_cached_total": 0, "words_failed_7d": 0,
    "topic_lists_total": 0, "topic_lists_7d": 0,
    "writing_lessons_total": 0,
    "writing_checks_total": 0, "writing_checks_7d": 0,
    "writing_reviews_total": 0, "writing_reviews_7d": 0,
    "input_tokens_30d": 0, "output_tokens_30d": 0
  },
  "writing": { "attempts_total": 0, "attempts_submitted": 0, "avg_band_30d": null },  // number | null
  "feedback": { "new": 0, "total": 0 },
  "content_reports": { "total": 0, "last_7d": 0, "needs_review": 0 },
  "daily": [ { "date": "2026-10-18", "signups": 0, "active_users": 0, "reviews": 0, "ai_requests": 0 } ], // 30 days, oldest first
  "recent_reports": [ { "id": "", "headword": "", "reason": "wrong_meaning", "note": "", "user_email": "", "created_at": "" } ]
}

// GET /api/v1/admin/users  → items[]
{
  "id": "", "email": "", "display_name": "", "avatar_url": "", "ui_locale": "vi",
  "created_at": "", "last_login_at": null, "last_active_at": null,
  "decks": 0, "cards": 0, "reviews_total": 0, "reviews_7d": 0,
  "ai_words": 0, "topic_lists": 0,
  "writing_attempts": 0, "writing_submitted": 0, "writing_ai_calls": 0,
  "feedback": 0
}

// GET /api/v1/admin/feedback → items[] (also the PATCH response)
{
  "id": "", "category": "bug", "message": "", "page": "/vi/decks", "ui_locale": "vi", "user_agent": "",
  "status": "new", "admin_note": "", "notified": true,
  "created_at": "", "updated_at": "", "resolved_at": null,
  "user": { "id": "", "email": "", "display_name": "", "avatar_url": "" }
}
```

AI requests per day = charged generation items + topic list requests + writing AI requests (lesson, check, submit). Tokens in 30 days = `generation_jobs` + `topic_suggestion_requests` + `writing_ai_requests`.

New error codes: `FEEDBACK_LIMIT`, `FEEDBACK_NOT_FOUND`, `ADMIN_UNAUTHORIZED`, `ADMIN_INVALID_CREDENTIALS`. Admin routes use the per-IP short-term rate limit, and login the medium-term one.

```sql
CREATE TABLE admin_accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(64) NOT NULL,
    password_hash TEXT NOT NULL,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_admin_accounts_username ON admin_accounts (lower(username));
```

## 9. Web

- `features/feedback`: `FeedbackDialog` (opened from `UserMenu`), `useSendFeedback`.
- `features/admin`: `adminClient` (fetches `/api/v1/admin/*`; on 401 goes to the admin login), views for Overview, Users and Feedback, React Query hooks.
- `app/[locale]/admin/login` and `app/[locale]/admin/(board)/{page,users,feedback}` with the admin layout; `lib/admin-actions.ts` (sign in, sign out).
- All text in `en.json` and `vi.json` (`feedback.*`, `admin.*`).

## 10. User stories

- **US-6.1** As a learner, I send feedback from any page, and I see that it was received.
- **US-6.2** As the owner, I get a Telegram message in our group for each new feedback, with a link to it.
- **US-6.3** As the owner, I sign in to the admin board with the account seeded from `be/.env`; nobody else can open it or call its API.
- **US-6.4** As the owner, I see the key numbers of the app on one page.
- **US-6.5** As the owner, I see every user with their activity and AI use, and find one by email.
- **US-6.6** As the owner, I read feedback, mark it seen or resolved, and keep a note.

## 11. Decisions

1. The admin account is seeded in the database and checked by the API (2026-11-16, replaces the first plan of credentials in `fe/.env` plus a shared API key: fewer secrets, no extra proxy route).
2. The admin session does not reuse the learner JWT: no `role` column on users, and a learner token can never reach the board.
3. Telegram is best effort (no retry queue). `notified_at` shows the failures on the board.
4. Board numbers are computed on request (small data). Add caching or materialised views when it gets slow.

## 12. Definition of done

- Integration tests: send feedback (validation, daily limit, notifier called with the right fields), admin seed (create, idempotent, password change), login, admin token checks (none / learner token / admin token), overview and users numbers on seeded data, feedback status changes.
- e2e: a learner sends feedback; the admin signs in, sees it, resolves it; wrong password is rejected.
- `make test` and `make e2e` pass; README documents the variables and the Telegram setup.
