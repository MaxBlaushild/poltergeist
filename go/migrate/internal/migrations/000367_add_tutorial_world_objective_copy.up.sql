ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS scenario_objective_copy TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS monster_objective_copy TEXT NOT NULL DEFAULT '';

UPDATE tutorial_configs
SET scenario_objective_copy = 'Complete the tutorial scenario to continue.'
WHERE TRIM(COALESCE(scenario_objective_copy, '')) = '';

UPDATE tutorial_configs
SET monster_objective_copy = 'Defeat the tutorial monster encounter to continue.'
WHERE TRIM(COALESCE(monster_objective_copy, '')) = '';
