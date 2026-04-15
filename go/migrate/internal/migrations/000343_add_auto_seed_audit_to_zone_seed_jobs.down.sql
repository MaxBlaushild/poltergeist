ALTER TABLE zone_seed_jobs
DROP COLUMN IF EXISTS auto_seed_audit,
DROP COLUMN IF EXISTS seed_mode;
