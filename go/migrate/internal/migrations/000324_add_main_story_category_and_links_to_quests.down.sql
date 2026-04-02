ALTER TABLE quests
DROP CONSTRAINT IF EXISTS quests_category_check;

ALTER TABLE quests
DROP COLUMN IF EXISTS main_story_next_quest_id,
DROP COLUMN IF EXISTS main_story_previous_quest_id,
DROP COLUMN IF EXISTS category;

ALTER TABLE quest_archetypes
DROP CONSTRAINT IF EXISTS quest_archetypes_category_check;

ALTER TABLE quest_archetypes
DROP COLUMN IF EXISTS category;
