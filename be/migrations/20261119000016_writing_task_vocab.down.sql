CREATE TABLE writing_lesson_vocab (
    lesson_id     UUID PRIMARY KEY REFERENCES writing_lessons (id) ON DELETE CASCADE,
    content       JSONB NOT NULL,
    model         VARCHAR(64) NOT NULL DEFAULT '',
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO writing_lesson_vocab (lesson_id, content, model, input_tokens, output_tokens, created_at)
SELECT l.id, v.content, v.model, v.input_tokens, v.output_tokens, v.created_at
FROM writing_task_vocab v JOIN writing_lessons l ON l.task_id = v.task_id;

DROP TABLE writing_task_vocab;
