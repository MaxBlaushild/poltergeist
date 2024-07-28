CREATE TABLE match_verification_codes (
    id UUID PRIMARY KEY,
    match_id UUID NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    verification_code_id UUID NOT NULL,
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (verification_code_id) REFERENCES verification_codes(id)
);

CREATE INDEX idx_match_verification_codes_match_id ON match_verification_codes(match_id);
CREATE INDEX idx_match_verification_codes_verification_code_id ON match_verification_codes(verification_code_id);
CREATE UNIQUE INDEX idx_match_verification_codes_unique ON match_verification_codes(match_id, verification_code_id);


