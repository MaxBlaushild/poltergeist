CREATE TABLE team_matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id),
    match_id UUID NOT NULL REFERENCES matches(id)
);
