ALTER TABLE quests
  DROP COLUMN IF EXISTS monster_encounter_target_level;

ALTER TABLE quest_archetypes
  DROP COLUMN IF EXISTS monster_encounter_target_level;
