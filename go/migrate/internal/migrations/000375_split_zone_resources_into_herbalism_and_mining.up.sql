ALTER TABLE zone_kinds
ADD COLUMN IF NOT EXISTS herbalism_resource_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
ADD COLUMN IF NOT EXISTS mining_resource_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1;

UPDATE zone_kinds
SET
  herbalism_resource_count_ratio = resource_count_ratio,
  mining_resource_count_ratio = resource_count_ratio;

ALTER TABLE zone_seed_jobs
ADD COLUMN IF NOT EXISTS herbalism_resource_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS mining_resource_count INTEGER NOT NULL DEFAULT 0;

UPDATE zone_seed_jobs
SET
  herbalism_resource_count = (resource_count + 1) / 2,
  mining_resource_count = resource_count / 2
WHERE resource_count > 0
  AND herbalism_resource_count = 0
  AND mining_resource_count = 0;
