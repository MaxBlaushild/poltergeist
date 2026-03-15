ALTER TABLE zone_seed_jobs
  DROP COLUMN IF EXISTS boss_encounter_count,
  DROP COLUMN IF EXISTS raid_encounter_count;
