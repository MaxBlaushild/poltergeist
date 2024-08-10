CREATE TABLE point_of_interest_group_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    point_of_interest_group_id UUID NOT NULL,
    point_of_interest_id UUID NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    FOREIGN KEY (point_of_interest_group_id) REFERENCES point_of_interest_groups(id),
    FOREIGN KEY (point_of_interest_id) REFERENCES point_of_interest(id)
);
