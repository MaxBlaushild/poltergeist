CREATE TABLE IF NOT EXISTS point_of_interest_discoveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    team_id UUID REFERENCES teams(id),
    user_id UUID REFERENCES users(id),
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id)
);
