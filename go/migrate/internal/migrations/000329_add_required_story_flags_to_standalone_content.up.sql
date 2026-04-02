ALTER TABLE scenarios
    ADD COLUMN IF NOT EXISTS required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE monster_encounters
    ADD COLUMN IF NOT EXISTS required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb;
