-- Create trending_destinations table
CREATE TABLE trending_destinations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    location_type VARCHAR NOT NULL CHECK (location_type IN ('city', 'country')),
    place_id VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    formatted_address TEXT NOT NULL,
    document_count INTEGER NOT NULL,
    rank INTEGER NOT NULL CHECK (rank >= 1 AND rank <= 5),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    UNIQUE(location_type, rank)
);

CREATE INDEX idx_trending_destinations_location_type ON trending_destinations(location_type);
CREATE INDEX idx_trending_destinations_place_id ON trending_destinations(place_id);
