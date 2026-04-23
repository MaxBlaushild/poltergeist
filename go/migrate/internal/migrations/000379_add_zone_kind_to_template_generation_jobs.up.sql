ALTER TABLE scenario_template_generation_jobs
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';

ALTER TABLE challenge_template_generation_jobs
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';
