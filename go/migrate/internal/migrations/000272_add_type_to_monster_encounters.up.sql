ALTER TABLE monster_encounters
  ADD COLUMN encounter_type TEXT NOT NULL DEFAULT 'monster';

UPDATE monster_encounters
SET encounter_type = 'monster'
WHERE encounter_type IS NULL OR BTRIM(encounter_type) = '';
