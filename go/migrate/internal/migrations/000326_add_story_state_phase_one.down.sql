ALTER TABLE points_of_interest
DROP COLUMN IF EXISTS story_variants;

ALTER TABLE characters
DROP COLUMN IF EXISTS story_variants;

ALTER TABLE quest_archetypes
DROP COLUMN IF EXISTS clear_story_flags,
DROP COLUMN IF EXISTS set_story_flags,
DROP COLUMN IF EXISTS required_story_flags;

ALTER TABLE quests
DROP COLUMN IF EXISTS clear_story_flags,
DROP COLUMN IF EXISTS set_story_flags,
DROP COLUMN IF EXISTS required_story_flags;

DROP INDEX IF EXISTS idx_user_story_flags_user_key;

DROP TABLE IF EXISTS user_story_flags;
