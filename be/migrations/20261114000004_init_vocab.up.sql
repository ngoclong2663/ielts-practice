-- Epic 3: smart card generation.

-- Shared, per-word generated content. One row per (headword, sense hint, meaning language).
CREATE TABLE lexemes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             TEXT NOT NULL UNIQUE,
    headword        VARCHAR(100) NOT NULL,
    lemma           VARCHAR(100) NOT NULL DEFAULT '',
    sense_hint      VARCHAR(60)  NOT NULL DEFAULT '',
    meaning_lang    VARCHAR(8)   NOT NULL DEFAULT 'vi',
    content         JSONB,
    content_version SMALLINT NOT NULL DEFAULT 1,
    -- pending: being generated · ready · invalid: not a word (suggestion set) · failed · needs_review: reported too often
    status          VARCHAR(16) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'ready', 'invalid', 'failed', 'needs_review')),
    suggestion      VARCHAR(100) NOT NULL DEFAULT '',
    image_query     VARCHAR(200) NOT NULL DEFAULT '',
    image_candidates JSONB NOT NULL DEFAULT '[]',
    quality_flags   JSONB NOT NULL DEFAULT '[]',
    report_count    INT NOT NULL DEFAULT 0,
    claimed_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE generation_jobs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_deck_id    UUID,
    status            VARCHAR(16) NOT NULL DEFAULT 'queued'
                      CHECK (status IN ('queued', 'running', 'done', 'partial', 'failed')),
    total             INT NOT NULL DEFAULT 0,
    done              INT NOT NULL DEFAULT 0,
    committed_deck_id UUID,
    input_tokens      INT NOT NULL DEFAULT 0,
    output_tokens     INT NOT NULL DEFAULT 0,
    cache_read_tokens INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at       TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_generation_jobs_user ON generation_jobs (user_id, created_at DESC);

CREATE TABLE generation_job_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES generation_jobs (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    position        INT NOT NULL,
    input_text      VARCHAR(100) NOT NULL,
    sense_hint      VARCHAR(60)  NOT NULL DEFAULT '',
    lexeme_id       UUID REFERENCES lexemes (id) ON DELETE SET NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'queued'
                    CHECK (status IN ('queued', 'content', 'image', 'ready', 'failed', 'invalid')),
    -- regenerate: skip the shared cache and keep the fresh result on this item only.
    regenerate      BOOLEAN NOT NULL DEFAULT FALSE,
    -- charged: this item caused an LLM call (counts toward the daily quota).
    charged         BOOLEAN NOT NULL DEFAULT FALSE,
    cached          BOOLEAN NOT NULL DEFAULT FALSE,
    -- The user's working copy of the card (edits and image choice before saving).
    content         JSONB,
    image_candidates JSONB NOT NULL DEFAULT '[]',
    error_code      VARCHAR(64) NOT NULL DEFAULT '',
    suggestion      VARCHAR(100) NOT NULL DEFAULT '',
    attempts        INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    removed         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_job_items_job ON generation_job_items (job_id, position);
CREATE INDEX idx_job_items_queue ON generation_job_items (next_attempt_at) WHERE status = 'queued';
CREATE INDEX idx_job_items_user_charged ON generation_job_items (user_id, created_at) WHERE charged;

-- Direct provider links that expire (Pixabay). Cards point at /api/v1/images/<provider>/<id>,
-- which redirects to the current link and refreshes it when it has expired.
CREATE TABLE image_links (
    provider     VARCHAR(16) NOT NULL,
    provider_id  VARCHAR(64) NOT NULL,
    url          TEXT NOT NULL DEFAULT '',
    thumb_url    TEXT NOT NULL DEFAULT '',
    expires_at   TIMESTAMPTZ NOT NULL,
    query        VARCHAR(200) NOT NULL DEFAULT '',
    headword     VARCHAR(100) NOT NULL DEFAULT '',
    -- Set when the photo was removed upstream and a replacement was found (a CardImage JSON).
    replacement  JSONB,
    gone         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, provider_id)
);

CREATE TABLE lexeme_reports (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lexeme_id  UUID NOT NULL REFERENCES lexemes (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reason     VARCHAR(32) NOT NULL,
    note       VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (lexeme_id, user_id)
);
