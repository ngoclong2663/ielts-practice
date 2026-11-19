-- Epic 8 §3: vocabulary library. Library cards are lexemes (the shared card cache of Epic 3) with origin
-- 'library': learner jobs reuse them but never overwrite them, and reports do not send them back to the AI.
ALTER TABLE lexemes ADD COLUMN origin VARCHAR(16) NOT NULL DEFAULT 'generated'
    CHECK (origin IN ('generated', 'library'));

CREATE TABLE library_collections (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(64)  NOT NULL UNIQUE,
    kind         VARCHAR(16)  NOT NULL,          -- ielts | toeic | ...
    title        JSONB        NOT NULL,          -- {"vi": "...", "en": "..."}
    description  JSONB        NOT NULL DEFAULT '{}',
    sort         SMALLINT     NOT NULL DEFAULT 0,
    status       VARCHAR(16)  NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'hidden')),
    content_hash CHAR(64)     NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE library_topics (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID        NOT NULL REFERENCES library_collections (id) ON DELETE CASCADE,
    topic_id      VARCHAR(32) NOT NULL,          -- Epic 4 topic id (vocab.topics.<id> on the web)
    sort          SMALLINT    NOT NULL DEFAULT 0,
    content_hash  CHAR(64)    NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (collection_id, topic_id)
);

CREATE TABLE library_words (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    library_topic_id UUID        NOT NULL REFERENCES library_topics (id) ON DELETE CASCADE,
    lexeme_id        UUID        NOT NULL REFERENCES lexemes (id),
    cefr             VARCHAR(4)  NOT NULL DEFAULT '',
    source           VARCHAR(16) NOT NULL DEFAULT 'core',   -- ielts4000 | core
    sort             SMALLINT    NOT NULL DEFAULT 0,
    -- headword, forms, synonyms and Vietnamese meaning, lowercased and without accents, for search
    search_text      TEXT        NOT NULL DEFAULT '',
    UNIQUE (library_topic_id, lexeme_id)
);
CREATE INDEX idx_library_words_lexeme ON library_words (lexeme_id);
