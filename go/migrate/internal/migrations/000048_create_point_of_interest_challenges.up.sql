CREATE TABLE point_of_interest_challenges (
    id UUID PRIMARY KEY,
    point_of_interest_id UUID NOT NULL,
    question TEXT NOT NULL,
    tier INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (point_of_interest_id) REFERENCES points_of_interest(id)
);
