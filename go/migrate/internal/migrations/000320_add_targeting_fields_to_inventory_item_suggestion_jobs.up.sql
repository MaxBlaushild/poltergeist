ALTER TABLE inventory_item_suggestion_jobs
  ADD COLUMN IF NOT EXISTS stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS benefit_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS status_names JSONB NOT NULL DEFAULT '[]'::jsonb;
