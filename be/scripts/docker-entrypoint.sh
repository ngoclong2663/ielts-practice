#!/bin/sh
set -e

# Apply pending migrations before starting the API (disable with RUN_MIGRATIONS=false).
if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
  if [ -z "$DATABASE_URL" ]; then
    echo "DATABASE_URL is required to run migrations" >&2
    exit 1
  fi
  echo "Running database migrations..."
  migrate -path /app/migrations -database "$DATABASE_URL" up
fi

# Apply the seed data compiled into the binary (writing task packs). Idempotent; disable with RUN_SEEDS=false.
if [ "${RUN_SEEDS:-true}" = "true" ]; then
  echo "Applying seed data..."
  api seed
fi

exec api
