CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE text_verification_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    phone_number VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL,
    used BOOLEAN NOT NULL
);