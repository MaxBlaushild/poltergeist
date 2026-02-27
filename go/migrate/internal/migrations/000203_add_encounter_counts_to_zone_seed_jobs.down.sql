ALTER TABLE zone_seed_jobs
  DROP COLUMN IF EXISTS input_encounter_count,
  DROP COLUMN IF EXISTS option_encounter_count;
