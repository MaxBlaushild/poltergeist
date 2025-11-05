CREATE TABLE IF NOT EXISTS user_treasure_chest_openings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id),
    treasure_chest_id UUID NOT NULL REFERENCES treasure_chests(id),
    UNIQUE(user_id, treasure_chest_id)
);

CREATE INDEX idx_user_treasure_chest_openings_user_id ON user_treasure_chest_openings(user_id);
CREATE INDEX idx_user_treasure_chest_openings_treasure_chest_id ON user_treasure_chest_openings(treasure_chest_id);

