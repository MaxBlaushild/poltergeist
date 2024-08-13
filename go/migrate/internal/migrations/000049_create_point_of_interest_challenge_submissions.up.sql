CREATE TABLE point_of_interest_challenge_submissions (
    id UUID PRIMARY KEY,
    point_of_interest_challenge_id UUID NOT NULL,
    team_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    text TEXT,
    image_url TEXT,
    is_correct BOOLEAN,
    FOREIGN KEY (point_of_interest_challenge_id) REFERENCES point_of_interest_challenges(id),
    FOREIGN KEY (team_id) REFERENCES teams(id)
);
