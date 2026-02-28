ALTER TABLE scenarios
ADD COLUMN success_reward_mode TEXT NOT NULL DEFAULT 'shared',
ADD COLUMN success_health_restore_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN success_health_restore_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN success_mana_restore_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN success_mana_restore_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN success_statuses JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE scenario_options
ADD COLUMN success_health_restore_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN success_health_restore_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN success_mana_restore_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN success_mana_restore_value INTEGER NOT NULL DEFAULT 0,
ADD COLUMN success_statuses JSONB NOT NULL DEFAULT '[]'::jsonb;
