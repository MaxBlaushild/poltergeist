CREATE TABLE social_accounts (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  account_id VARCHAR(128),
  username VARCHAR(128),
  access_token TEXT NOT NULL,
  refresh_token TEXT,
  expires_at TIMESTAMPTZ,
  scopes TEXT
);

CREATE UNIQUE INDEX idx_social_accounts_user_provider ON social_accounts(user_id, provider);
CREATE INDEX idx_social_accounts_user_id ON social_accounts(user_id);
