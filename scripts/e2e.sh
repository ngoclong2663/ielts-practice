#!/usr/bin/env bash
# End-to-end tests: test-mode API (:8081) + isolated web build (:3001) + Playwright.
# Uses its own database (<your DB>_e2e) and build folder, so it never touches `make be` / `make fe`.
# Usage: make e2e            (screenshots: E2E_SCREENSHOTS=/some/dir make e2e)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_PORT="${E2E_API_PORT:-8081}"
WEB_PORT="${E2E_WEB_PORT:-3001}"

# Database: be/.env DATABASE_URL with the database name suffixed by _e2e (override with E2E_DATABASE_URL).
if [ -z "${E2E_DATABASE_URL:-}" ]; then
  BASE_URL="$(grep -E '^DATABASE_URL=' "$ROOT/be/.env" | cut -d= -f2-)"
  E2E_DATABASE_URL="$(echo "$BASE_URL" | sed -E 's#/([A-Za-z0-9_]+)(\?|$)#/\1_e2e\2#')"
fi
DB_NAME="$(echo "$E2E_DATABASE_URL" | sed -E 's#.*/([A-Za-z0-9_]+)(\?.*)?$#\1#')"
ADMIN_URL="$(echo "$E2E_DATABASE_URL" | sed -E "s#/${DB_NAME}#/postgres#")"
psql "$ADMIN_URL" -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" 2>/dev/null | grep -q 1 \
  || psql "$ADMIN_URL" -c "CREATE DATABASE ${DB_NAME}" >/dev/null 2>&1 \
  || docker exec postgres psql -U postgres -c "CREATE DATABASE ${DB_NAME}" >/dev/null 2>&1 || true
(cd "$ROOT/be" && DATABASE_URL="$E2E_DATABASE_URL" bash scripts/migrate-up.sh >/dev/null)

# Admin board (Epic 6): `api seed` creates this admin account in the e2e database.
export E2E_ADMIN_USERNAME="${E2E_ADMIN_USERNAME:-owner}"
export E2E_ADMIN_PASSWORD="${E2E_ADMIN_PASSWORD:-e2e-admin-password}"

for port in "$API_PORT" "$WEB_PORT"; do
  if curl -s -o /dev/null "http://localhost:$port"; then
    echo "Port $port is already in use; stop that process first." >&2
    exit 1
  fi
done

# Each server runs in its own process group so the whole tree (pnpm -> next-server) is stopped.
# `set -m` gives every background job its own group (pid = pgid) on Linux and macOS (no setsid there).
set -m
cleanup() {
  for pid in "${API_PID:-}" "${WEB_PID:-}"; do
    [ -n "$pid" ] && kill -- "-$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

echo "Starting test-mode API on :$API_PORT"
(cd "$ROOT/be" && go build -o bin/api ./cmd)
cd "$ROOT/be"
# Writing task bank (figures are not uploaded: S3 is off in e2e) and the admin account.
APP_ENV=test DATABASE_URL="$E2E_DATABASE_URL" S3_BUCKET_NAME= \
  ADMIN_USERNAME="$E2E_ADMIN_USERNAME" ADMIN_PASSWORD="$E2E_ADMIN_PASSWORD" ./bin/api seed >/dev/null
APP_ENV=test PORT="$API_PORT" DATABASE_URL="$E2E_DATABASE_URL" FRONTEND_URL="http://localhost:$WEB_PORT" \
  RATE_LIMIT_DISABLED=true S3_BUCKET_NAME= VOCAB_FAKE_PROVIDERS=true \
  TELEGRAM_BOT_TOKEN= TELEGRAM_CHAT_ID= ./bin/api >/dev/null 2>&1 &
API_PID=$!

echo "Building web app into .next-e2e"
(cd "$ROOT/fe" && NEXT_DIST_DIR=.next-e2e API_URL="http://localhost:$API_PORT" pnpm -s build >/dev/null)
cd "$ROOT/fe"
NEXT_DIST_DIR=.next-e2e API_URL="http://localhost:$API_PORT" pnpm -s next start -p "$WEB_PORT" >/dev/null 2>&1 &
WEB_PID=$!

for _ in $(seq 1 60); do
  curl -sf "http://localhost:$API_PORT/api/v1/health" >/dev/null && curl -sf -o /dev/null "http://localhost:$WEB_PORT/en" && break
  sleep 1
done

cd "$ROOT/fe" && E2E_BASE_URL="http://localhost:$WEB_PORT" E2E_API_URL="http://localhost:$API_PORT" pnpm exec playwright test "$@"
