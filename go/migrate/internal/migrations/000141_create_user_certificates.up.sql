CREATE TABLE user_certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL UNIQUE,
    certificate BYTEA NOT NULL,
    certificate_pem TEXT NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint BYTEA NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_certificates_user_id ON user_certificates(user_id);
CREATE INDEX idx_user_certificates_fingerprint ON user_certificates(fingerprint);
