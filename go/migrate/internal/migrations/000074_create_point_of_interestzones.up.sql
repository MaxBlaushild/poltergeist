CREATE TABLE point_of_interest_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    zone_id UUID NOT NULL REFERENCES zones(id),
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id)
);

CREATE INDEX idx_point_of_interest_zones_deleted_at ON point_of_interest_zones(deleted_at);
