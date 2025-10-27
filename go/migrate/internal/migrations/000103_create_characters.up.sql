BEGIN;

CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    map_icon_url VARCHAR(500),
    dialogue_image_url VARCHAR(500),
    location_id UUID REFERENCES points_of_interest(id),
    movement_pattern_id UUID REFERENCES movement_patterns(id) UNIQUE
);

CREATE INDEX idx_characters_name ON characters(name);
CREATE INDEX idx_characters_location_id ON characters(location_id);
CREATE INDEX idx_characters_movement_pattern_id ON characters(movement_pattern_id);

COMMIT;
