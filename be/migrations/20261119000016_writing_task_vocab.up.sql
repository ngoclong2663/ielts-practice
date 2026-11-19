-- Epic 8 §4: useful vocabulary per task instead of per lesson (it is the same for every position and attempt),
-- seeded and reviewed. Vocabulary already written for a lesson is kept as its task's (the first one written).
CREATE TABLE writing_task_vocab (
    task_id       UUID PRIMARY KEY REFERENCES writing_tasks (id) ON DELETE CASCADE,
    content       JSONB NOT NULL,
    model         VARCHAR(64) NOT NULL DEFAULT '',   -- 'seed' for reviewed seed files
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    content_hash  CHAR(64) NOT NULL DEFAULT '',      -- seed file hash, so an unchanged file is skipped
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO writing_task_vocab (task_id, content, model, input_tokens, output_tokens, created_at)
SELECT DISTINCT ON (l.task_id) l.task_id, v.content, v.model, v.input_tokens, v.output_tokens, v.created_at
FROM writing_lesson_vocab v JOIN writing_lessons l ON l.id = v.lesson_id
ORDER BY l.task_id, v.created_at;

DROP TABLE writing_lesson_vocab;
