CREATE TABLE user_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    effect TEXT NOT NULL DEFAULT '',
    effect_type TEXT NOT NULL DEFAULT 'stat_modifier',
    strength_mod INTEGER NOT NULL DEFAULT 0,
    dexterity_mod INTEGER NOT NULL DEFAULT 0,
    constitution_mod INTEGER NOT NULL DEFAULT 0,
    intelligence_mod INTEGER NOT NULL DEFAULT 0,
    wisdom_mod INTEGER NOT NULL DEFAULT 0,
    charisma_mod INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_statuses_user_id_expires_at ON user_statuses(user_id, expires_at);
CREATE INDEX idx_user_statuses_active_stat_modifiers ON user_statuses(user_id, effect_type, started_at, expires_at);
