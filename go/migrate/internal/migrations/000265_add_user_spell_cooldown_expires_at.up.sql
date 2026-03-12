ALTER TABLE user_spells
  ADD COLUMN IF NOT EXISTS cooldown_expires_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_user_spells_cooldown_expires_at
  ON user_spells(cooldown_expires_at);
