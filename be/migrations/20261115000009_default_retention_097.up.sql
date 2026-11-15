-- New learners start with 97% desired retention: short review gaps for new exam vocabulary
-- (Good each time: 10m → 1d → 2d → 4d → 7d → 13d → 22d instead of 10m → 2d → 11d → 46d → 163d at 90%).
-- Existing accounts keep the value they have.
ALTER TABLE user_study_settings ALTER COLUMN desired_retention SET DEFAULT 0.97;
