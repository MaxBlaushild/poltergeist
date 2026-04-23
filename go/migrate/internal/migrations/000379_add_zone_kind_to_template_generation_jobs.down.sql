ALTER TABLE challenge_template_generation_jobs
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE scenario_template_generation_jobs
  DROP COLUMN IF EXISTS zone_kind;
