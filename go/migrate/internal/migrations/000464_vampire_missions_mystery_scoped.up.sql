-- Missions become mystery-scoped too, the same way secrets already are:
-- the same character cast in two different mysteries/subplots can have a
-- different set of missions in each. Nullable indefinitely, same posture
-- as vampire_secrets.mystery_id — a mission with no mystery is simply
-- unusable by any instance (never served to a player), rather than a hard
-- constraint blocking edge cases.
ALTER TABLE vampire_missions ADD COLUMN IF NOT EXISTS mystery_id UUID REFERENCES vampire_mysteries(id);
CREATE INDEX IF NOT EXISTS idx_vampire_missions_mystery_id ON vampire_missions(mystery_id);
CREATE INDEX IF NOT EXISTS idx_vampire_missions_character_mystery ON vampire_missions(character_id, mystery_id);

-- Backfill: every existing mission moves onto the legacy mystery (the same
-- one 000461 created/found for quiz questions, secrets, and instances).
DO $$
DECLARE
  legacy_mystery_id UUID;
BEGIN
  SELECT id INTO legacy_mystery_id FROM vampire_mysteries WHERE name = 'The Crimson Toast' LIMIT 1;
  IF legacy_mystery_id IS NOT NULL THEN
    UPDATE vampire_missions SET mystery_id = legacy_mystery_id WHERE mystery_id IS NULL;
  END IF;
END $$;
