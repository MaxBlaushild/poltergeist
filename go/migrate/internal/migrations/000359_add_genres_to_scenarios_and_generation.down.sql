BEGIN;

DROP INDEX IF EXISTS idx_scenario_template_generation_jobs_genre_id;
DROP INDEX IF EXISTS idx_scenario_generation_jobs_genre_id;
DROP INDEX IF EXISTS idx_scenario_templates_genre_id;
DROP INDEX IF EXISTS idx_scenarios_genre_id;

ALTER TABLE scenario_template_generation_jobs
  DROP CONSTRAINT IF EXISTS scenario_template_generation_jobs_genre_id_fkey;

ALTER TABLE scenario_generation_jobs
  DROP CONSTRAINT IF EXISTS scenario_generation_jobs_genre_id_fkey;

ALTER TABLE scenario_templates
  DROP CONSTRAINT IF EXISTS scenario_templates_genre_id_fkey;

ALTER TABLE scenarios
  DROP CONSTRAINT IF EXISTS scenarios_genre_id_fkey;

ALTER TABLE scenario_template_generation_jobs
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE scenario_generation_jobs
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE scenario_templates
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS genre_id;

COMMIT;
