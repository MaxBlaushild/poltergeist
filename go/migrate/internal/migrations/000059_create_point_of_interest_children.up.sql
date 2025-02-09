CREATE TABLE point_of_interest_children (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    point_of_interest_group_member_id UUID NOT NULL REFERENCES point_of_interest_group_members(id),
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id),
    point_of_interest_challenge_id UUID NOT NULL REFERENCES point_of_interest_challenges(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
