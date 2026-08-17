-- Subplots: a sibling of mysteries — same shape (summary, full lore, beats
-- whose beats are tied to character secrets the exact same way), but an
-- instance can pick zero or many of them in addition to its one required
-- mystery. Reusing vampire_mysteries/vampire_mystery_beats/vampire_secrets
-- wholesale (via this new is_subplot flag) rather than a parallel table set
-- — a subplot is structurally just a mystery row, so every existing
-- beat/secret/eligibility mechanism already works unmodified against a
-- subplot's id. The one deliberate difference: subplots don't have quiz
-- questions (an instance's quiz is always just its main mystery's), and
-- selecting a subplot doesn't narrow invite eligibility the way the main
-- mystery does — both by design (see chat decisions).
ALTER TABLE vampire_mysteries ADD COLUMN IF NOT EXISTS is_subplot BOOLEAN NOT NULL DEFAULT FALSE;

-- An instance's subplots — zero or many, alongside its one required
-- mystery_id. No ordinal/status; a subplot is either selected or not.
CREATE TABLE IF NOT EXISTS vampire_instance_subplots (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  mystery_id UUID NOT NULL REFERENCES vampire_mysteries(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, mystery_id)
);
CREATE INDEX IF NOT EXISTS idx_vampire_instance_subplots_mystery_id ON vampire_instance_subplots(mystery_id);
