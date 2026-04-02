ALTER TABLE district_seed_jobs
ADD COLUMN IF NOT EXISTS zone_seed_settings jsonb NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS zone_seed_job_ids jsonb NOT NULL DEFAULT '[]'::jsonb;
