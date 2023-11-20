CREATE TABLE crystal_unlockings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    team_id UUID NOT NULL,
    crystal_id UUID NOT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id),
    FOREIGN KEY (crystal_id) REFERENCES crystals(id)
);