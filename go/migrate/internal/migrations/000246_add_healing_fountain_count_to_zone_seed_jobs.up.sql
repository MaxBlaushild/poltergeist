ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS healing_fountain_count INTEGER NOT NULL DEFAULT 0;
