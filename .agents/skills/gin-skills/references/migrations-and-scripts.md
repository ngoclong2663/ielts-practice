# Go / Gin Migrations & Scripts Reference

This guide explains the SQL migration structure, migration scripts, and Makefile helpers used in the Go / Gin project.

---

## 🗄️ SQL Migration Structure (`migrations/`)

Migrations are stored as pairs of raw SQL scripts timestamped in UTC:

```text
migrations/
├── 20260403163646_init_users_table.up.sql
└── 20260403163646_init_users_table.down.sql
```

### Up Script Example (`.up.sql`)

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### Down Script Example (`.down.sql`)

```sql
DROP TABLE IF EXISTS users;
```

---

## 📜 Shell Scripts (`scripts/`)

Migration runners live in the `scripts/` directory:

- `scripts/migrate-up.sh`: Executes all pending `*.up.sql` files against the configured database.
- `scripts/migrate-down.sh`: Rolls back the last applied migration.
- `scripts/migrate-upto.sh <version>`: Migrates up to a specific version timestamp.
- `scripts/migrate-downto.sh <version>`: Rolls back down to a specific version timestamp.

---

## ⌨️ Makefile Commands

```bash
# Run server with live reload
make run

# Run all pending migrations
make mgup

# Rollback last migration
make mgdown

# Generate new migration files
make newmg name="create_orders_table"
```
