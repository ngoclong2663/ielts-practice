ALTER TABLE topic_suggestions ADD COLUMN level VARCHAR(8) NOT NULL DEFAULT 'b2' CHECK (level IN ('b1', 'b2', 'c1'));
