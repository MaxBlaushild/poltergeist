DROP INDEX IF EXISTS idx_posts_cert_fingerprint;
DROP INDEX IF EXISTS idx_posts_manifest_hash;

ALTER TABLE posts DROP COLUMN IF EXISTS asset_id;
ALTER TABLE posts DROP COLUMN IF EXISTS cert_fingerprint;
ALTER TABLE posts DROP COLUMN IF EXISTS manifest_uri;
ALTER TABLE posts DROP COLUMN IF EXISTS manifest_hash;
