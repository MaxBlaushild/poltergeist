CREATE TABLE points_of_interest (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    name TEXT NOT NULL,
    clue TEXT NOT NULL,
    capture_challenge TEXT NOT NULL,
    attune_challenge TEXT NOT NULL,
    lat TEXT NOT NULL,
    lng TEXT NOT NULL,
    image_url TEXT
);

