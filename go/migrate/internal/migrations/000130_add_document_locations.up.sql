-- Create document_locations table
CREATE TABLE document_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    document_id UUID NOT NULL,
    place_id VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    formatted_address TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location_type VARCHAR NOT NULL CHECK (location_type IN ('city', 'country', 'continent')),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE INDEX idx_document_locations_document_id ON document_locations(document_id);
CREATE INDEX idx_document_locations_place_id ON document_locations(place_id);
CREATE INDEX idx_document_locations_location_type ON document_locations(location_type);

