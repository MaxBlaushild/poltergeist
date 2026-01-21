ALTER TABLE user_certificates ADD COLUMN active BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_user_certificates_active ON user_certificates(active);
