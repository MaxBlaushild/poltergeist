CREATE TABLE verification_codes (
    id UUID PRIMARY KEY,
    code CHAR(6) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX idx_verification_codes_code ON verification_codes(code);
