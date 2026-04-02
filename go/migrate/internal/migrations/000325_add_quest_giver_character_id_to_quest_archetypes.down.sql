DROP INDEX IF EXISTS idx_quest_archetypes_quest_giver_character_id;

ALTER TABLE quest_archetypes
DROP COLUMN IF EXISTS quest_giver_character_id;
