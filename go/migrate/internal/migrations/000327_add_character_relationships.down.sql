ALTER TABLE quest_archetypes
    DROP COLUMN IF EXISTS quest_giver_relationship_effects;

ALTER TABLE quests
    DROP COLUMN IF EXISTS quest_giver_relationship_effects;

DROP INDEX IF EXISTS user_character_relationships_user_character_idx;

DROP TABLE IF EXISTS user_character_relationships;
