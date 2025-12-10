-- Create quick_decision_requests table
CREATE TABLE quick_decision_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    question TEXT NOT NULL,
    option_1 VARCHAR NOT NULL,
    option_2 VARCHAR NOT NULL,
    option_3 VARCHAR,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_quick_decision_requests_user_id ON quick_decision_requests(user_id);
