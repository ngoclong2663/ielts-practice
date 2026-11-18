DELETE FROM writing_ai_requests WHERE kind = 'vocab';
ALTER TABLE writing_ai_requests DROP CONSTRAINT writing_ai_requests_kind_check;
ALTER TABLE writing_ai_requests ADD CONSTRAINT writing_ai_requests_kind_check
    CHECK (kind IN ('lesson', 'check', 'submit'));

DROP TABLE IF EXISTS writing_lesson_vocab;
