CREATE TABLE neighboring_points_of_interest (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    point_of_interest_one_id UUID NOT NULL REFERENCES points_of_interest(id),
    point_of_interest_two_id UUID NOT NULL REFERENCES points_of_interest(id)
);
