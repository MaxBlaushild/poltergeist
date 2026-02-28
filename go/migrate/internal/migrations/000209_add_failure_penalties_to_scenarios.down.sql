ALTER TABLE scenario_options
DROP COLUMN IF EXISTS failure_statuses,
DROP COLUMN IF EXISTS failure_mana_drain_value,
DROP COLUMN IF EXISTS failure_mana_drain_type,
DROP COLUMN IF EXISTS failure_health_drain_value,
DROP COLUMN IF EXISTS failure_health_drain_type;

ALTER TABLE scenarios
DROP COLUMN IF EXISTS failure_statuses,
DROP COLUMN IF EXISTS failure_mana_drain_value,
DROP COLUMN IF EXISTS failure_mana_drain_type,
DROP COLUMN IF EXISTS failure_health_drain_value,
DROP COLUMN IF EXISTS failure_health_drain_type,
DROP COLUMN IF EXISTS failure_penalty_mode;
