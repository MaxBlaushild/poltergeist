DROP INDEX IF EXISTS exposition_template_generation_jobs_zone_kind_idx;
DROP INDEX IF EXISTS exposition_template_generation_jobs_created_at_idx;
DROP TABLE IF EXISTS exposition_template_generation_jobs;

ALTER TABLE zone_seed_jobs
  DROP COLUMN IF EXISTS exposition_count;

DROP INDEX IF EXISTS idx_expositions_exposition_template_id;

ALTER TABLE expositions
  DROP COLUMN IF EXISTS exposition_template_id;

ALTER TABLE exposition_templates
  DROP COLUMN IF EXISTS zone_kind;
