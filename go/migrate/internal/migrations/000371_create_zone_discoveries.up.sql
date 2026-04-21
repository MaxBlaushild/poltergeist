CREATE TABLE IF NOT EXISTS zone_discoveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    CONSTRAINT zone_discoveries_user_zone_unique UNIQUE (user_id, zone_id)
);

CREATE INDEX IF NOT EXISTS idx_zone_discoveries_user_id
ON zone_discoveries(user_id);

CREATE INDEX IF NOT EXISTS idx_zone_discoveries_zone_id
ON zone_discoveries(zone_id);
