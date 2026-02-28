ALTER TABLE scenarios
ADD COLUMN failure_penalty_mode TEXT NOT NULL DEFAULT 'shared',
ADD COLUMN failure_health_drain_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN failure_health_drain_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN failure_mana_drain_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN failure_mana_drain_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN failure_statuses JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE scenario_options
ADD COLUMN failure_health_drain_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN failure_health_drain_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN failure_mana_drain_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN failure_mana_drain_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN failure_statuses JSONB NOT NULL DEFAULT '[]'::jsonb;
