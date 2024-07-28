CREATE TABLE point_of_interest_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id),
    sonar_activity_id UUID NOT NULL REFERENCES sonar_activities(id)
);
