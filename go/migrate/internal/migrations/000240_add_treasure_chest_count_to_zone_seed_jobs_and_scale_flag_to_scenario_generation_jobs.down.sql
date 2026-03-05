ALTER TABLE scenario_generation_jobs
  DROP COLUMN IF EXISTS scale_with_user_level;

ALTER TABLE zone_seed_jobs
  DROP COLUMN IF EXISTS treasure_chest_count;
