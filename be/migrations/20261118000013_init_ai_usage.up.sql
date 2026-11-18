-- Epic 7: one row per call to a language model, written by the llm module for every model it tries
-- (a fallback answer costs the failed primary call too). The admin board prices it per model.
CREATE TABLE ai_usage (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider          VARCHAR(16) NOT NULL,              -- anthropic | google
    model             VARCHAR(64) NOT NULL,
    feature           VARCHAR(32) NOT NULL DEFAULT '',   -- vocab.cards, vocab.topic_words, writing.lesson, ...
    ok                BOOLEAN NOT NULL,                  -- false: error, refusal or cut off (tokens may still be billed)
    input_tokens      INT NOT NULL DEFAULT 0,            -- uncached input
    output_tokens     INT NOT NULL DEFAULT 0,            -- includes thinking tokens
    cache_read_tokens INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_usage_created ON ai_usage (created_at);

-- Writing calls recorded their model before this table existed; keep them so the board starts with history.
-- (Card generation and topic lists did not record a model, so they start from this migration.)
INSERT INTO ai_usage (provider, model, feature, ok, input_tokens, output_tokens, created_at)
SELECT CASE WHEN model LIKE 'gemini%' THEN 'google' ELSE 'anthropic' END,
       model, 'writing.' || kind, true, input_tokens, output_tokens, created_at
FROM writing_ai_requests
WHERE model <> '' AND model <> 'fake';
