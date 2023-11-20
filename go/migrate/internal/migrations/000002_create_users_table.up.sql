CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    name VARCHAR(255),
    phone_number VARCHAR(255) UNIQUE,
    active BOOLEAN
);