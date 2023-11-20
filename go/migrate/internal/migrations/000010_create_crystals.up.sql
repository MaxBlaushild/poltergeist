CREATE TABLE crystals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name VARCHAR(255) NOT NULL,
    clue TEXT NOT NULL,
    capture_challenge TEXT NOT NULL,
    attune_challenge TEXT NOT NULL,
    captured BOOLEAN NOT NULL,
    attuned BOOLEAN NOT NULL,
    lat VARCHAR(255),
    lng VARCHAR(255),
    capture_team_id UUID,
    FOREIGN KEY (capture_team_id) REFERENCES teams(id)
);