DROP TABLE IF EXISTS writing_ai_requests;
DROP TABLE IF EXISTS writing_submissions;
DROP TABLE IF EXISTS writing_attempt_paragraphs;
DROP TABLE IF EXISTS writing_attempts;
DROP TABLE IF EXISTS writing_lessons;
ALTER TABLE writing_tasks DROP COLUMN IF EXISTS positions;
