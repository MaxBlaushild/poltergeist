DROP INDEX IF EXISTS idx_user_spells_cooldown_expires_at;

ALTER TABLE user_spells
  DROP COLUMN IF EXISTS cooldown_expires_at;
