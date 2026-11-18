-- Epic 7: useful vocabulary for a writing lesson (topic words, paraphrases, task language), each with
-- synonyms and an example. Written once per lesson on first request and shared by every learner.
CREATE TABLE writing_lesson_vocab (
    lesson_id     UUID PRIMARY KEY REFERENCES writing_lessons (id) ON DELETE CASCADE,
    content       JSONB NOT NULL,
    model         VARCHAR(64) NOT NULL DEFAULT '',
    input_tokens  INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE writing_ai_requests DROP CONSTRAINT writing_ai_requests_kind_check;
ALTER TABLE writing_ai_requests ADD CONSTRAINT writing_ai_requests_kind_check
    CHECK (kind IN ('lesson', 'vocab', 'check', 'submit'));
