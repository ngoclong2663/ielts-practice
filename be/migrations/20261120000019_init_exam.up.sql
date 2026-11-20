-- Epic 10: IELTS Reading and Listening practice. Books, tests and parts are loaded by `api seed` from
-- seeds/exam/<book>/. The answer key sits in its own table so a content query can never return it.
CREATE TABLE exam_books (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(32) NOT NULL UNIQUE,                -- 's1'
    source_ref   VARCHAR(32) NOT NULL DEFAULT '',            -- internal: the book that served as reference
    title        JSONB NOT NULL,                             -- {"vi": "...", "en": "..."}
    sort         SMALLINT NOT NULL DEFAULT 0,
    status       VARCHAR(16) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'draft')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE exam_tests (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id    UUID NOT NULL REFERENCES exam_books (id) ON DELETE CASCADE,
    test_no    SMALLINT NOT NULL,
    slug       VARCHAR(32) NOT NULL UNIQUE,                  -- 's1-t1'
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (book_id, test_no)
);

CREATE TABLE exam_parts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id        UUID NOT NULL REFERENCES exam_tests (id) ON DELETE CASCADE,
    skill          VARCHAR(16) NOT NULL CHECK (skill IN ('reading', 'listening')),
    part_no        SMALLINT NOT NULL,
    slug           VARCHAR(48) NOT NULL UNIQUE,              -- 's1-t1-r1'
    source_ref     VARCHAR(64) NOT NULL DEFAULT '',          -- internal only
    title          VARCHAR(200) NOT NULL,
    topic          VARCHAR(32) NOT NULL DEFAULT '',
    q_from         SMALLINT NOT NULL,
    q_to           SMALLINT NOT NULL,
    question_types JSONB NOT NULL DEFAULT '[]',               -- group types, for filters and stats
    content        JSONB NOT NULL,                           -- passage or transcript and question groups, no key
    audio          JSONB,                                    -- listening: {key, duration, cues}
    status         VARCHAR(16) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'draft')),
    content_hash   CHAR(64) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (test_id, skill, part_no),
    CHECK (q_from >= 1 AND q_to <= 40 AND q_from <= q_to)
);

CREATE TABLE exam_keys (
    part_id    UUID PRIMARY KEY REFERENCES exam_parts (id) ON DELETE CASCADE,
    key        JSONB NOT NULL,                               -- {"1": {answer, evidence, pairs, why, trap}}
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A submitted run. Runs in progress (answers, flags, highlights, timer) stay in the learner's browser.
CREATE TABLE exam_attempts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    skill             VARCHAR(16) NOT NULL CHECK (skill IN ('reading', 'listening')),
    test_id           UUID NOT NULL REFERENCES exam_tests (id),
    part_ids          JSONB NOT NULL,                         -- ["<part id>", ...]
    kind              VARCHAR(8) NOT NULL CHECK (kind IN ('part', 'full', 'redo')),
    parent_attempt_id UUID REFERENCES exam_attempts (id) ON DELETE SET NULL,
    mode              VARCHAR(16) NOT NULL DEFAULT 'practice' CHECK (mode IN ('practice', 'exam')),
    time_limit_sec    INTEGER NOT NULL DEFAULT 0,             -- 0 = no limit
    duration_sec      INTEGER NOT NULL DEFAULT 0,
    questions         JSONB NOT NULL,                         -- the question numbers in this run
    answers           JSONB NOT NULL,
    results           JSONB NOT NULL,                         -- [{n, part_no, type, given, correct}]
    part_scores       JSONB NOT NULL,                         -- {"<part id>": {"correct": 9, "total": 13}}
    correct           SMALLINT NOT NULL,
    total             SMALLINT NOT NULL,
    band              NUMERIC(2, 1),                          -- full tests only
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exam_attempts_user ON exam_attempts (user_id, skill, created_at DESC);
CREATE INDEX idx_exam_parts_test ON exam_parts (test_id, skill, part_no);
