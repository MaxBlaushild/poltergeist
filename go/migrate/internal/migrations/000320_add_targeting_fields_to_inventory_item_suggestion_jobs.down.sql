ALTER TABLE inventory_item_suggestion_jobs
  DROP COLUMN IF EXISTS status_names,
  DROP COLUMN IF EXISTS benefit_tags,
  DROP COLUMN IF EXISTS stat_tags;
