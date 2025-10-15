CREATE TABLE activities (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    activity_type TEXT NOT NULL,
    seen BOOLEAN NOT NULL DEFAULT FALSE,
    data JSONB NOT NULL
);
