ALTER TABLE scenario_options
  DROP COLUMN IF EXISTS failure_text,
  DROP COLUMN IF EXISTS success_text;
