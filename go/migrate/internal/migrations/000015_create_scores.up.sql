BEGIN;

CREATE TABLE scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    score INTEGER NOT NULL
);

CREATE INDEX idx_scores_username ON scores(username);

COMMIT;