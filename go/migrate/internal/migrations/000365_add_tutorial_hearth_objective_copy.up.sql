ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS hearth_objective_copy TEXT NOT NULL DEFAULT '';

UPDATE tutorial_configs
SET hearth_objective_copy = 'Use your hearth to heal yourself before the tutorial continues.'
WHERE TRIM(COALESCE(hearth_objective_copy, '')) = '';
