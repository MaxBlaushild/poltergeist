BEGIN;

DROP INDEX IF EXISTS idx_user_base_daily_state_user_resets_on;
DROP INDEX IF EXISTS idx_user_base_structures_user_id;
DROP INDEX IF EXISTS idx_base_structure_level_costs_definition_level;
DROP INDEX IF EXISTS idx_base_structure_definitions_active_sort;
DROP INDEX IF EXISTS idx_base_resource_ledger_user_created_at;

DROP TABLE IF EXISTS user_base_daily_state;
DROP TABLE IF EXISTS user_base_structures;
DROP TABLE IF EXISTS base_structure_level_costs;
DROP TABLE IF EXISTS base_structure_definitions;
DROP TABLE IF EXISTS base_resource_ledger;
DROP TABLE IF EXISTS base_resource_balances;

COMMIT;
