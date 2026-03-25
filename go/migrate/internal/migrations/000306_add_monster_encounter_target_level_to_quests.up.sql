ALTER TABLE quest_archetypes
  ADD COLUMN IF NOT EXISTS monster_encounter_target_level INTEGER NOT NULL DEFAULT 1;

ALTER TABLE quests
  ADD COLUMN IF NOT EXISTS monster_encounter_target_level INTEGER NOT NULL DEFAULT 1;

UPDATE quest_archetypes
SET monster_encounter_target_level = GREATEST(1, COALESCE(difficulty, 1));

UPDATE quests
SET monster_encounter_target_level = GREATEST(1, COALESCE(difficulty, 1));
