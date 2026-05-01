BEGIN;

DROP TABLE IF EXISTS user_shrine_uses;
DROP TABLE IF EXISTS shrines;
DROP TABLE IF EXISTS shrine_template_generation_jobs;
DROP TABLE IF EXISTS shrine_templates;

ALTER TABLE zone_seed_jobs
  DROP COLUMN IF EXISTS shrine_count;

ALTER TABLE zone_kinds
  DROP COLUMN IF EXISTS shrine_count_ratio;

COMMIT;
