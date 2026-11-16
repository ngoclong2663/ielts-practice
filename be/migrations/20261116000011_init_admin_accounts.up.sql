-- Epic 6: the admin board account(s). Seeded by `api seed` from ADMIN_USERNAME / ADMIN_PASSWORD.
CREATE TABLE admin_accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(64) NOT NULL,
    password_hash TEXT NOT NULL, -- bcrypt
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_admin_accounts_username ON admin_accounts (lower(username));
