CREATE TABLE how_many_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    subscribed BOOLEAN DEFAULT false,
    num_free_questions INTEGER DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);