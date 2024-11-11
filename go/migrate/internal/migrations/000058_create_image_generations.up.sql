CREATE TABLE image_generations (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    generation_id VARCHAR(255) NOT NULL,
    generation_backend_id INTEGER NOT NULL,
    status INTEGER NOT NULL,
    option_one TEXT,
    option_two TEXT, 
    option_three TEXT,
    option_four TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
