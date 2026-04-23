ALTER TABLE scenario_template_generation_jobs
  ADD COLUMN IF NOT EXISTS yeet_it BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE monster_template_suggestion_jobs
  ADD COLUMN IF NOT EXISTS yeet_it BOOLEAN NOT NULL DEFAULT FALSE;
