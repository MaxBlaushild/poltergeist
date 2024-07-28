CREATE TABLE match_points_of_interest (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    match_id UUID NOT NULL REFERENCES matches(id),
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id)
);
