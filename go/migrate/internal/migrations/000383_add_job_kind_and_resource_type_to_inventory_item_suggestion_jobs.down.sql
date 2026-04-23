DROP INDEX IF EXISTS idx_inventory_item_suggestion_jobs_resource_type_id;
DROP INDEX IF EXISTS idx_inventory_item_suggestion_jobs_job_kind_created_at;

ALTER TABLE inventory_item_suggestion_jobs
  DROP CONSTRAINT IF EXISTS inventory_item_suggestion_jobs_resource_type_id_fkey;

ALTER TABLE inventory_item_suggestion_jobs
  DROP COLUMN IF EXISTS resource_type_id,
  DROP COLUMN IF EXISTS job_kind;
