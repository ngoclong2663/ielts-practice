-- Decks group a user's cards. Limits are NULL when the user's defaults apply.
CREATE TABLE decks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name            VARCHAR(80)  NOT NULL,
    description     VARCHAR(500) NOT NULL DEFAULT '',
    source          VARCHAR(16)  NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'paste', 'collection')),
    collection_id   UUID,
    new_per_day     INT CHECK (new_per_day BETWEEN 0 AND 500),
    reviews_per_day INT CHECK (reviews_per_day BETWEEN 0 AND 9999),
    archived_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_decks_user_name_active ON decks (user_id, lower(name)) WHERE archived_at IS NULL;
CREATE INDEX idx_decks_user ON decks (user_id) WHERE archived_at IS NULL;

-- Cards hold a content snapshot (JSON, see Epic 3 §4) plus the FSRS scheduling state.
-- state: 0 new, 1 learning, 2 review, 3 relearning (same values as go-fsrs).
CREATE TABLE cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    deck_id         UUID NOT NULL REFERENCES decks (id) ON DELETE CASCADE,
    lexeme_id       UUID,
    headword        VARCHAR(100) NOT NULL,
    content         JSONB NOT NULL,
    content_version SMALLINT NOT NULL DEFAULT 1,

    state           SMALLINT NOT NULL DEFAULT 0 CHECK (state BETWEEN 0 AND 3),
    due             TIMESTAMPTZ NOT NULL DEFAULT now(),
    position        INT NOT NULL DEFAULT 0,
    stability       DOUBLE PRECISION NOT NULL DEFAULT 0,
    difficulty      DOUBLE PRECISION NOT NULL DEFAULT 0,
    scheduled_days  INT NOT NULL DEFAULT 0,
    remaining_steps SMALLINT NOT NULL DEFAULT 0,
    reps            INT NOT NULL DEFAULT 0,
    lapses          INT NOT NULL DEFAULT 0,
    last_review_at  TIMESTAMPTZ,
    suspended       BOOLEAN NOT NULL DEFAULT FALSE,
    is_leech        BOOLEAN NOT NULL DEFAULT FALSE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cards_queue ON cards (user_id, state, due) WHERE NOT suspended;
CREATE INDEX idx_cards_deck_position ON cards (deck_id, position);
CREATE INDEX idx_cards_deck_headword ON cards (deck_id, lower(headword));
CREATE UNIQUE INDEX uq_cards_deck_lexeme ON cards (deck_id, lexeme_id) WHERE lexeme_id IS NOT NULL;

-- One row per answer. *_before columns let us undo exactly and later optimise FSRS weights.
CREATE TABLE review_logs (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    card_id                UUID NOT NULL REFERENCES cards (id) ON DELETE CASCADE,
    deck_id                UUID NOT NULL,
    client_review_id       UUID NOT NULL,
    rating                 SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 4),

    state_before           SMALLINT NOT NULL,
    due_before             TIMESTAMPTZ NOT NULL,
    stability_before       DOUBLE PRECISION NOT NULL,
    difficulty_before      DOUBLE PRECISION NOT NULL,
    scheduled_days_before  INT NOT NULL,
    remaining_steps_before SMALLINT NOT NULL,
    reps_before            INT NOT NULL,
    lapses_before          INT NOT NULL,
    last_review_before     TIMESTAMPTZ,
    is_leech_before        BOOLEAN NOT NULL,

    state_after            SMALLINT NOT NULL,
    due_after              TIMESTAMPTZ NOT NULL,
    stability_after        DOUBLE PRECISION NOT NULL,
    difficulty_after       DOUBLE PRECISION NOT NULL,
    scheduled_days_after   INT NOT NULL,
    elapsed_days           INT NOT NULL,

    duration_ms            INT NOT NULL DEFAULT 0,
    reviewed_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    undone_at              TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_review_logs_client ON review_logs (user_id, client_review_id);
CREATE INDEX idx_review_logs_user_time ON review_logs (user_id, reviewed_at DESC);
CREATE INDEX idx_review_logs_card ON review_logs (card_id, reviewed_at);

ALTER TABLE user_study_settings
    ADD COLUMN max_interval_days INT NOT NULL DEFAULT 36500 CHECK (max_interval_days BETWEEN 1 AND 36500);
