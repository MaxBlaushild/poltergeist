-- Player invites become character-agnostic — instead of a Host assigning a
-- specific character at invite time, an invitee picks their own from the
-- instance's curated pool after accepting (see
-- vampire_instance_character_pool and the self-select character picker).
-- Which character a person will play is no longer known until they've
-- accepted and chosen.
ALTER TABLE vampire_player_invites DROP COLUMN IF EXISTS character_id;

-- The pool of characters a Host has made selectable by players in this
-- Toast, curated from the full mystery-eligible set (the same "has secrets
-- for this instance's mystery" rule that used to gate who could be invited
-- at all). A character not in the pool can't be self-selected even if
-- otherwise eligible.
CREATE TABLE IF NOT EXISTS vampire_instance_character_pool (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, character_id)
);
CREATE INDEX IF NOT EXISTS idx_vampire_instance_character_pool_character_id
  ON vampire_instance_character_pool(character_id);

-- Backfill: every existing instance's pool starts as every character
-- eligible for its mystery (the same "everyone eligible" set that was
-- implicitly offerable before this feature), so nothing already invitable
-- disappears for a Host who hasn't visited the new Character Pool tab yet.
INSERT INTO vampire_instance_character_pool (instance_id, character_id)
SELECT DISTINCT i.id, s.character_id
FROM vampire_instances i
JOIN vampire_secrets s ON s.mystery_id = i.mystery_id
JOIN vampire_characters c ON c.id = s.character_id AND c.role_type = 'player'
ON CONFLICT DO NOTHING;
