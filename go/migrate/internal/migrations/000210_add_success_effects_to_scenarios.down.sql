ALTER TABLE scenario_options
DROP COLUMN IF EXISTS success_statuses,
DROP COLUMN IF EXISTS success_mana_restore_value,
DROP COLUMN IF EXISTS success_mana_restore_type,
DROP COLUMN IF EXISTS success_health_restore_value,
DROP COLUMN IF EXISTS success_health_restore_type;

ALTER TABLE scenarios
DROP COLUMN IF EXISTS success_statuses,
DROP COLUMN IF EXISTS success_mana_restore_value,
DROP COLUMN IF EXISTS success_mana_restore_type,
DROP COLUMN IF EXISTS success_health_restore_value,
DROP COLUMN IF EXISTS success_health_restore_type,
DROP COLUMN IF EXISTS success_reward_mode;
