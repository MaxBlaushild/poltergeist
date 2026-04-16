ALTER TABLE scenario_templates
  ADD COLUMN IF NOT EXISTS success_handoff_text TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS failure_handoff_text TEXT NOT NULL DEFAULT '';

ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS success_handoff_text TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS failure_handoff_text TEXT NOT NULL DEFAULT '';

ALTER TABLE scenario_options
  ADD COLUMN IF NOT EXISTS success_handoff_text TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS failure_handoff_text TEXT NOT NULL DEFAULT '';
