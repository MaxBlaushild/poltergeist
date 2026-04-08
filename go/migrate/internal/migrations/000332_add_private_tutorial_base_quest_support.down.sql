DROP INDEX IF EXISTS idx_quests_owner_user_id;
ALTER TABLE quests
  DROP COLUMN IF EXISTS ephemeral,
  DROP COLUMN IF EXISTS owner_user_id;

DROP INDEX IF EXISTS idx_characters_owner_user_id;
ALTER TABLE characters
  DROP COLUMN IF EXISTS ephemeral,
  DROP COLUMN IF EXISTS owner_user_id;

ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS base_quest_giver_character_id,
  DROP COLUMN IF EXISTS base_quest_archetype_id;
