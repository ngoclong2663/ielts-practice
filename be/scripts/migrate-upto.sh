#!/bin/bash

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    # Explicit environment variables win over .env (e.g. DATABASE_URL=... make mgup).
    while IFS='=' read -r key value; do
        case "$key" in ''|\#*) continue ;; esac
        [ -z "${!key+x}" ] && export "$key=$value"
    done < .env
fi

# Default to local database if DB_URL is not set
DB_URL="${DATABASE_URL:-${DB_URL:-postgresql://postgres:postgres@localhost:5432/ielts_practice?sslmode=disable}}"

# Path to migrations directory
MIGRATIONS_PATH="migrations"
VERSION="$1"

if [ -z "$VERSION" ]; then
    echo "Usage: bash scripts/migrate-upto.sh <version>"
    exit 1
fi

echo "Migrating to version ${VERSION}..."
echo "Database URL: $(echo "$DB_URL" | sed -E 's#(://[^:]+:)[^@]+@#\1****@#')"

# Go to a specific migration version
migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" goto "$VERSION"

if [ $? -eq 0 ]; then
    echo "Migration completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi
