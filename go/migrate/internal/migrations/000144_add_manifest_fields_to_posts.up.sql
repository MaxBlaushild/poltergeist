ALTER TABLE posts ADD COLUMN manifest_hash BYTEA;
ALTER TABLE posts ADD COLUMN manifest_uri TEXT;
ALTER TABLE posts ADD COLUMN cert_fingerprint BYTEA;
ALTER TABLE posts ADD COLUMN asset_id TEXT;

CREATE INDEX idx_posts_manifest_hash ON posts(manifest_hash);
CREATE INDEX idx_posts_cert_fingerprint ON posts(cert_fingerprint);
