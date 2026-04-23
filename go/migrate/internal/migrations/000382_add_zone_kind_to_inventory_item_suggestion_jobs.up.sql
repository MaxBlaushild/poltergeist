ALTER TABLE inventory_item_suggestion_jobs
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';
