ALTER TABLE scenario_options
  DROP COLUMN IF EXISTS success_handoff_text,
  DROP COLUMN IF EXISTS failure_handoff_text;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS success_handoff_text,
  DROP COLUMN IF EXISTS failure_handoff_text;

ALTER TABLE scenario_templates
  DROP COLUMN IF EXISTS success_handoff_text,
  DROP COLUMN IF EXISTS failure_handoff_text;
