-- Per-character sigil (PIN) used to validate a player landed on the right
-- character before any content is shown.
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS password TEXT NOT NULL DEFAULT '';
