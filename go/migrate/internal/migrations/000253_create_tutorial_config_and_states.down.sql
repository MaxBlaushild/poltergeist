DROP INDEX IF EXISTS idx_user_tutorial_states_scenario_id;
DROP TABLE IF EXISTS user_tutorial_states;
DROP TABLE IF EXISTS tutorial_configs;
DROP INDEX IF EXISTS idx_scenarios_owner_user_id;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS ephemeral,
  DROP COLUMN IF EXISTS owner_user_id;
