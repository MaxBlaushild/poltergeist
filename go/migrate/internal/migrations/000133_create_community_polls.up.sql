-- Create community_polls table
CREATE TABLE community_polls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    question TEXT NOT NULL,
    options JSONB NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (jsonb_array_length(options) >= 3 AND jsonb_array_length(options) <= 10)
);

CREATE INDEX idx_community_polls_user_id ON community_polls(user_id);
