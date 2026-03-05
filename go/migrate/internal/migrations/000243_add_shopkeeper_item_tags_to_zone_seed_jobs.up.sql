ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS shopkeeper_item_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
