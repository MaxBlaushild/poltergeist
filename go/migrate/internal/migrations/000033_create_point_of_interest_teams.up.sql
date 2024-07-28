CREATE TABLE point_of_interest_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    team_id UUID NOT NULL REFERENCES teams(id),
    point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id),
    unlocked BOOLEAN NOT NULL DEFAULT false,
    captured BOOLEAN NOT NULL DEFAULT false,
    attuned BOOLEAN NOT NULL DEFAULT false
);
