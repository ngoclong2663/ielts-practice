-- Epic 5: the writing task bank. Rows come from the seed packs in be/seeds/writing (`api seed`).
CREATE TABLE writing_tasks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(64) NOT NULL UNIQUE,         -- e.g. 'cam21-test1-task1'
    pack         VARCHAR(32) NOT NULL,                -- e.g. 'cam21'
    source_ref   VARCHAR(64) NOT NULL DEFAULT '',     -- which book task inspired ours, shown as "Inspired by …"
    test_no      SMALLINT NOT NULL DEFAULT 0,
    task_no      SMALLINT NOT NULL CHECK (task_no IN (1, 2)),
    module       VARCHAR(16) NOT NULL DEFAULT 'academic' CHECK (module IN ('academic', 'general')),
    task_type    VARCHAR(32) NOT NULL,
    topic        VARCHAR(32) NOT NULL DEFAULT '',
    title        VARCHAR(160) NOT NULL,
    prompt       TEXT NOT NULL,                       -- the task-specific text; standard instructions come from the UI
    -- [{key, sha256, content_type, width, height, alt}]; key is the S3 object key
    figures      JSONB NOT NULL DEFAULT '[]',
    -- the figure as data (series, table, stages, map changes) for lessons and feedback
    facts        JSONB NOT NULL DEFAULT '{}',
    min_words    SMALLINT NOT NULL,
    minutes      SMALLINT NOT NULL,
    status       VARCHAR(16) NOT NULL DEFAULT 'published' CHECK (status IN ('draft', 'published', 'archived')),
    content_hash CHAR(64) NOT NULL,                   -- SHA-256 of the seed entry and its figures
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_writing_tasks_list ON writing_tasks (task_no, status, task_type);
