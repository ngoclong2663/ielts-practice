-- Epic 4: words by topic.

-- Words a user said they already know: never suggested again.
CREATE TABLE user_known_words (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    lemma      VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lemma)
);

-- One word list for a topic; "Suggest more" appends to `words`.
CREATE TABLE topic_suggestions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    deck_id          UUID,
    topic_id         VARCHAR(32) NOT NULL DEFAULT '',
    topic_text       VARCHAR(60) NOT NULL DEFAULT '',
    level            VARCHAR(8)  NOT NULL CHECK (level IN ('b1', 'b2', 'c1')),
    requested        INT NOT NULL,
    -- [{lemma, pos, meaning_vi, cefr}] in the order shown
    words            JSONB NOT NULL DEFAULT '[]',
    skipped_existing INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_topic_suggestions_user ON topic_suggestions (user_id, created_at DESC);

-- One row per successful AI request (a new list or "Suggest more"): the daily limit and token usage.
CREATE TABLE topic_suggestion_requests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suggestion_id UUID NOT NULL REFERENCES topic_suggestions (id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    words_added   INT NOT NULL,
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_topic_suggestion_requests_user ON topic_suggestion_requests (user_id, created_at);
