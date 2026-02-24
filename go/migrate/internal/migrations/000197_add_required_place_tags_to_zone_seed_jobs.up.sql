ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS required_place_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
