CREATE TABLE how_many_questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    text TEXT NOT NULL,
    how_many INTEGER NOT NULL,
    explanation TEXT NOT NULL,
    valid BOOLEAN NOT NULL,
    done BOOLEAN NOT NULL
);