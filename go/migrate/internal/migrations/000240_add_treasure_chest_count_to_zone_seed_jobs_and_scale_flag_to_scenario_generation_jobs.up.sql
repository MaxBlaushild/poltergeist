ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS treasure_chest_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE scenario_generation_jobs
  ADD COLUMN IF NOT EXISTS scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE;
