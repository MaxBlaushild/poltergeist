CREATE TABLE tracked_point_of_interest_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    point_of_interest_group_id UUID NOT NULL REFERENCES point_of_interest_groups(id),
    user_id UUID NOT NULL REFERENCES users(id),
    UNIQUE(point_of_interest_group_id, user_id)
);

CREATE INDEX tracked_point_of_interest_groups_deleted_at_idx ON tracked_point_of_interest_groups(deleted_at);
CREATE INDEX tracked_point_of_interest_groups_user_id_idx ON tracked_point_of_interest_groups(user_id);
CREATE INDEX tracked_point_of_interest_groups_poi_group_id_idx ON tracked_point_of_interest_groups(point_of_interest_group_id);
