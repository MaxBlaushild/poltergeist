BEGIN;

CREATE TABLE sent_texts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    text_type VARCHAR(255) NOT NULL,
    phone_number VARCHAR(255) NOT NULL,
    text TEXT NOT NULL
);

CREATE INDEX idx_sent_texts_text_type ON sent_texts(text_type);
CREATE INDEX idx_sent_texts_phone_number ON sent_texts(phone_number);

COMMIT;