-- Epic 6: learner feedback, posted to Telegram and handled on the admin board.
CREATE TABLE feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category    VARCHAR(16)  NOT NULL CHECK (category IN ('bug', 'idea', 'content', 'other')),
    message     TEXT         NOT NULL CHECK (char_length(message) BETWEEN 1 AND 2000),
    page        VARCHAR(300) NOT NULL DEFAULT '',
    ui_locale   VARCHAR(8)   NOT NULL DEFAULT '',
    user_agent  VARCHAR(400) NOT NULL DEFAULT '',
    status      VARCHAR(16)  NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'seen', 'resolved')),
    admin_note  VARCHAR(1000) NOT NULL DEFAULT '',
    -- set when the Telegram message was delivered; NULL shows "not delivered" on the board
    notified_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feedback_status ON feedback (status, created_at DESC);
CREATE INDEX idx_feedback_user ON feedback (user_id, created_at);
