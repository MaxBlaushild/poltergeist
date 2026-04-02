ALTER TABLE monster_encounters
    DROP COLUMN IF EXISTS required_story_flags;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS required_story_flags;

ALTER TABLE scenarios
    DROP COLUMN IF EXISTS required_story_flags;
