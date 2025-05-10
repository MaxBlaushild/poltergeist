CREATE TABLE IF NOT EXISTS user_zone_reputations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    zone_id UUID NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    total_reputation INTEGER NOT NULL DEFAULT 0,
    reputation_on_level INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id),
    UNIQUE(user_id, zone_id)
);
