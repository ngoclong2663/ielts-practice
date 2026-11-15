-- Epic 5: guided writing practice.

-- Task 2 positions a learner can take (e.g. ["agree","disagree","partly"]); empty = the default for the task type.
ALTER TABLE writing_tasks ADD COLUMN positions JSONB NOT NULL DEFAULT '[]';

-- One lesson per task, position and task version: plan questions, paragraphs with Vietnamese
-- sentence prompts, hints and model sentences. Shared by every learner (Epic 5 §5).
CREATE TABLE writing_lessons (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id       UUID NOT NULL REFERENCES writing_tasks (id) ON DELETE CASCADE,
    position      VARCHAR(24) NOT NULL DEFAULT '',
    task_hash     CHAR(64) NOT NULL,                  -- writing_tasks.content_hash it was written for
    content       JSONB NOT NULL,
    model         VARCHAR(64) NOT NULL DEFAULT '',
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, position, task_hash)
);

-- A learner's run through one task.
CREATE TABLE writing_attempts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    task_id      UUID NOT NULL REFERENCES writing_tasks (id) ON DELETE CASCADE,
    lesson_id    UUID NOT NULL REFERENCES writing_lessons (id),
    level        SMALLINT NOT NULL CHECK (level BETWEEN 1 AND 4),
    position     VARCHAR(24) NOT NULL DEFAULT '',
    plan_answers JSONB,                               -- chosen option per plan question; NULL until answered
    status       VARCHAR(16) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'submitted')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at TIMESTAMPTZ
);

CREATE INDEX idx_writing_attempts_user ON writing_attempts (user_id, created_at DESC);
CREATE INDEX idx_writing_attempts_task ON writing_attempts (user_id, task_id);

-- The latest check of one paragraph: the learner's sentences and the feedback per sentence.
CREATE TABLE writing_attempt_paragraphs (
    attempt_id    UUID NOT NULL REFERENCES writing_attempts (id) ON DELETE CASCADE,
    paragraph_key VARCHAR(16) NOT NULL,
    answers       JSONB NOT NULL,
    feedback      JSONB NOT NULL,
    checked_count INT NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (attempt_id, paragraph_key)
);

-- The full answer and its feedback (one per attempt).
CREATE TABLE writing_submissions (
    attempt_id    UUID PRIMARY KEY REFERENCES writing_attempts (id) ON DELETE CASCADE,
    text          TEXT NOT NULL,
    word_count    INT NOT NULL,
    seconds_spent INT NOT NULL DEFAULT 0,
    feedback      JSONB NOT NULL,
    band_low      NUMERIC(2, 1) NOT NULL,
    band_high     NUMERIC(2, 1) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per AI call made for a learner: daily limits (check, submit) and cost tracking (all kinds).
CREATE TABLE writing_ai_requests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    attempt_id    UUID REFERENCES writing_attempts (id) ON DELETE SET NULL,
    kind          VARCHAR(16) NOT NULL CHECK (kind IN ('lesson', 'check', 'submit')),
    model         VARCHAR(64) NOT NULL DEFAULT '',
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_writing_ai_requests_user ON writing_ai_requests (user_id, kind, created_at);
