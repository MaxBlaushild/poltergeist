CREATE TABLE quest_acceptances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id),
    point_of_interest_group_id UUID NOT NULL REFERENCES point_of_interest_groups(id),
    character_id UUID NOT NULL REFERENCES characters(id),
    UNIQUE(user_id, point_of_interest_group_id)
);

CREATE INDEX idx_quest_acceptances_user_id ON quest_acceptances(user_id);
CREATE INDEX idx_quest_acceptances_point_of_interest_group_id ON quest_acceptances(point_of_interest_group_id);

