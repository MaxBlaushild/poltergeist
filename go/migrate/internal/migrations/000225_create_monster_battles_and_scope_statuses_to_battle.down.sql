DROP INDEX IF EXISTS idx_monster_statuses_battle_id_active_stat_modifiers;
DROP INDEX IF EXISTS idx_monster_statuses_battle_id_expires_at;

ALTER TABLE monster_statuses
DROP CONSTRAINT IF EXISTS fk_monster_statuses_battle_id;

ALTER TABLE monster_statuses
DROP COLUMN IF EXISTS battle_id;

CREATE INDEX idx_monster_statuses_user_monster_expires_at
ON monster_statuses(user_id, monster_id, expires_at);

CREATE INDEX idx_monster_statuses_user_monster_active_stat_modifiers
ON monster_statuses(user_id, monster_id, effect_type, started_at, expires_at);

DROP INDEX IF EXISTS idx_monster_battles_user_monster_active;
DROP TABLE IF EXISTS monster_battles;
