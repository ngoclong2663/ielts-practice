-- Epic 7: short guide videos (how to learn vocabulary, how the writing levels work). The MP4 and its poster
-- live in S3 under content-addressed keys; `api seed` loads this table from seeds/guides/videos.json.
CREATE TABLE guide_videos (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             VARCHAR(64) NOT NULL UNIQUE,
    topic            VARCHAR(16) NOT NULL CHECK (topic IN ('vocab', 'writing')),
    sort             SMALLINT NOT NULL DEFAULT 0,
    lang             VARCHAR(8) NOT NULL DEFAULT 'vi',          -- language of the narration
    title            JSONB NOT NULL,                            -- {"vi": "...", "en": "..."}
    description      JSONB NOT NULL,
    video_key        VARCHAR(200) NOT NULL,                     -- guides/videos/<sha256>.mp4
    poster_key       VARCHAR(200) NOT NULL DEFAULT '',          -- guides/posters/<sha256>.jpg
    duration_seconds NUMERIC(6, 2) NOT NULL,
    status           VARCHAR(16) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'draft')),
    content_hash     CHAR(64) NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_guide_videos_topic ON guide_videos (topic, sort);
