DROP INDEX IF EXISTS idx_monster_statuses_user_monster_active_stat_modifiers;
DROP INDEX IF EXISTS idx_monster_statuses_user_monster_expires_at;

ALTER TABLE monster_statuses
DROP CONSTRAINT IF EXISTS fk_monster_statuses_user_id;

ALTER TABLE monster_statuses
DROP COLUMN IF EXISTS user_id;

CREATE INDEX idx_monster_statuses_monster_id_expires_at
ON monster_statuses(monster_id, expires_at);

CREATE INDEX idx_monster_statuses_active_stat_modifiers
ON monster_statuses(monster_id, effect_type, started_at, expires_at);
