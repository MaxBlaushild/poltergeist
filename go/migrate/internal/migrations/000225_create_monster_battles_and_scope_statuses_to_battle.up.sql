CREATE TABLE monster_battles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    monster_id UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (monster_id) REFERENCES monsters(id) ON DELETE CASCADE
);

CREATE INDEX idx_monster_battles_user_monster_active
ON monster_battles(user_id, monster_id, started_at DESC)
WHERE ended_at IS NULL;

ALTER TABLE monster_statuses
ADD COLUMN battle_id UUID;

DELETE FROM monster_statuses;

ALTER TABLE monster_statuses
ALTER COLUMN battle_id SET NOT NULL;

ALTER TABLE monster_statuses
ADD CONSTRAINT fk_monster_statuses_battle_id
FOREIGN KEY (battle_id) REFERENCES monster_battles(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_monster_statuses_user_monster_expires_at;
DROP INDEX IF EXISTS idx_monster_statuses_user_monster_active_stat_modifiers;

CREATE INDEX idx_monster_statuses_battle_id_expires_at
ON monster_statuses(battle_id, expires_at);

CREATE INDEX idx_monster_statuses_battle_id_active_stat_modifiers
ON monster_statuses(battle_id, effect_type, started_at, expires_at);
