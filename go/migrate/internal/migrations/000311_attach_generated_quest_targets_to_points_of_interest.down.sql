DROP INDEX IF EXISTS monster_encounters_point_of_interest_id_idx;

ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS point_of_interest_id;
