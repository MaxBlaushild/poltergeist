BEGIN;

CREATE TABLE movement_patterns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    movement_pattern_type VARCHAR(255) NOT NULL CHECK (movement_pattern_type IN ('static', 'random', 'path')),
    zone_id UUID REFERENCES zones(id),
    starting_latitude DOUBLE PRECISION NOT NULL,
    starting_longitude DOUBLE PRECISION NOT NULL,
    path JSONB DEFAULT '[]'::jsonb
);

CREATE INDEX idx_movement_patterns_zone_id ON movement_patterns(zone_id);
CREATE INDEX idx_movement_patterns_type ON movement_patterns(movement_pattern_type);

COMMIT;
