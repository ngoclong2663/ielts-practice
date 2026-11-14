#!/bin/bash

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Default to local database if DB_URL is not set
DB_URL="${DATABASE_URL:-${DB_URL:-postgresql://postgres:postgres@localhost:5432/ielts_practice?sslmode=disable}}"

# Path to migrations directory
MIGRATIONS_PATH="migrations"
STEPS="$1"

echo "Running migrations down..."
echo "Database URL: $(echo "$DB_URL" | sed -E 's#(://[^:]+:)[^@]+@#\1****@#')"

# Run migrations
if [ -n "$STEPS" ]; then
    migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" down "$STEPS"
else
    migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" down
fi

if [ $? -eq 0 ]; then
    echo "Migrations rolled back successfully!"
else
    echo "Migration rollback failed!"
    exit 1
fi