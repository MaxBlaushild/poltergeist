ALTER TABLE scenario_options
  ADD COLUMN success_text TEXT NOT NULL DEFAULT '',
  ADD COLUMN failure_text TEXT NOT NULL DEFAULT '';
