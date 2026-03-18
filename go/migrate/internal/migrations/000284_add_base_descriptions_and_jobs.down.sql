DROP INDEX IF EXISTS base_description_generation_jobs_created_at_idx;
DROP INDEX IF EXISTS base_description_generation_jobs_status_idx;
DROP INDEX IF EXISTS base_description_generation_jobs_base_id_idx;
DROP TABLE IF EXISTS base_description_generation_jobs;

ALTER TABLE bases
DROP COLUMN IF EXISTS description;
