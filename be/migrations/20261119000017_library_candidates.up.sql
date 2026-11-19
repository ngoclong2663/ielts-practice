-- Epic 9: library candidates. AI-made cards that several learners keep are offered on the admin board; the
-- owner approves one into a library topic (it becomes a library word, source 'learner', its card locked) or
-- rejects it (not offered again).
CREATE TABLE library_candidate_decisions (
    lexeme_id  UUID PRIMARY KEY REFERENCES lexemes (id) ON DELETE CASCADE,
    decision   VARCHAR(16) NOT NULL CHECK (decision IN ('approved', 'rejected')),
    topic_id   VARCHAR(32) NOT NULL DEFAULT '',
    decided_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
