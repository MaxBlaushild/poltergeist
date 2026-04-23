ALTER TABLE inventory_item_suggestion_jobs
  ADD COLUMN IF NOT EXISTS job_kind TEXT NOT NULL DEFAULT 'draft_batch',
  ADD COLUMN IF NOT EXISTS resource_type_id UUID;

UPDATE inventory_item_suggestion_jobs
SET job_kind = 'draft_batch'
WHERE COALESCE(TRIM(job_kind), '') = '';

ALTER TABLE inventory_item_suggestion_jobs
  DROP CONSTRAINT IF EXISTS inventory_item_suggestion_jobs_resource_type_id_fkey,
  ADD CONSTRAINT inventory_item_suggestion_jobs_resource_type_id_fkey
    FOREIGN KEY (resource_type_id) REFERENCES resource_types(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_item_suggestion_jobs_job_kind_created_at
  ON inventory_item_suggestion_jobs (job_kind, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_inventory_item_suggestion_jobs_resource_type_id
  ON inventory_item_suggestion_jobs (resource_type_id);
