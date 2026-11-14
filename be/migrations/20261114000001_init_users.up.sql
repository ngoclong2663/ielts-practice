CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub     VARCHAR(255) NOT NULL UNIQUE,
    email          VARCHAR(320) NOT NULL,
    display_name   VARCHAR(120) NOT NULL DEFAULT '',
    avatar_url     TEXT NOT NULL DEFAULT '',
    ui_locale      VARCHAR(8)  NOT NULL DEFAULT 'en' CHECK (ui_locale IN ('vi', 'en')),
    timezone       VARCHAR(64) NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
    day_start_hour SMALLINT    NOT NULL DEFAULT 4 CHECK (day_start_hour BETWEEN 0 AND 23),
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users (lower(email));

CREATE TABLE user_study_settings (
    user_id             UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    new_cards_per_day   INT          NOT NULL DEFAULT 20  CHECK (new_cards_per_day BETWEEN 0 AND 500),
    max_reviews_per_day INT          NOT NULL DEFAULT 200 CHECK (max_reviews_per_day BETWEEN 0 AND 9999),
    desired_retention   NUMERIC(3,2) NOT NULL DEFAULT 0.90 CHECK (desired_retention BETWEEN 0.70 AND 0.99),
    learning_steps      VARCHAR(64)  NOT NULL DEFAULT '1m 10m',
    relearning_steps    VARCHAR(64)  NOT NULL DEFAULT '10m',
    audio_autoplay      BOOLEAN      NOT NULL DEFAULT TRUE,
    accent              VARCHAR(2)   NOT NULL DEFAULT 'uk' CHECK (accent IN ('uk', 'us')),
    show_image_on_front BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);
