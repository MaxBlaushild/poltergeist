ALTER TABLE zone_seed_jobs
DROP COLUMN IF EXISTS mining_resource_count,
DROP COLUMN IF EXISTS herbalism_resource_count;

ALTER TABLE zone_kinds
DROP COLUMN IF EXISTS mining_resource_count_ratio,
DROP COLUMN IF EXISTS herbalism_resource_count_ratio;
