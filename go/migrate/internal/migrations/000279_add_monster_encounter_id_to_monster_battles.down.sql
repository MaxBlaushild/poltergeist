DROP INDEX IF EXISTS monster_battles_monster_encounter_id_idx;

ALTER TABLE monster_battles
  DROP COLUMN IF EXISTS monster_encounter_id;
