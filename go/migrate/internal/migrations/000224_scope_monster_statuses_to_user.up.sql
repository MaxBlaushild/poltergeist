ALTER TABLE monster_statuses
ADD COLUMN user_id UUID;

DELETE FROM monster_statuses;

ALTER TABLE monster_statuses
ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE monster_statuses
ADD CONSTRAINT fk_monster_statuses_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_monster_statuses_monster_id_expires_at;
DROP INDEX IF EXISTS idx_monster_statuses_active_stat_modifiers;

CREATE INDEX idx_monster_statuses_user_monster_expires_at
ON monster_statuses(user_id, monster_id, expires_at);

CREATE INDEX idx_monster_statuses_user_monster_active_stat_modifiers
ON monster_statuses(user_id, monster_id, effect_type, started_at, expires_at);
