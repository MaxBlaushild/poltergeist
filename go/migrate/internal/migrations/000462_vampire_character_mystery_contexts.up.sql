-- A character's post-Act-1 context becomes mystery-scoped too, same
-- rationale as secrets: the same character cast in two different mysteries
-- knows a different version of "what happens after Act One" in each. One
-- row per (character, mystery) — unlike secrets this is a single string,
-- not a list, so it's upserted rather than wholesale-replaced.
CREATE TABLE IF NOT EXISTS vampire_character_mystery_contexts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  mystery_id UUID NOT NULL REFERENCES vampire_mysteries(id) ON DELETE CASCADE,
  post_act1_context TEXT NOT NULL DEFAULT '',
  UNIQUE (character_id, mystery_id)
);
CREATE INDEX IF NOT EXISTS idx_vampire_character_mystery_contexts_mystery_id
  ON vampire_character_mystery_contexts(mystery_id);

-- Backfill: every character's existing post_act1_context moves to a row
-- scoped to the legacy mystery (same one the 000461 migration created/found
-- for quiz questions, secrets, and instances).
DO $$
DECLARE
  legacy_mystery_id UUID;
BEGIN
  SELECT id INTO legacy_mystery_id FROM vampire_mysteries WHERE name = 'The Crimson Toast' LIMIT 1;
  IF legacy_mystery_id IS NOT NULL THEN
    INSERT INTO vampire_character_mystery_contexts (character_id, mystery_id, post_act1_context)
    SELECT id, legacy_mystery_id, post_act1_context
    FROM vampire_characters
    WHERE post_act1_context <> ''
    ON CONFLICT (character_id, mystery_id) DO NOTHING;
  END IF;
END $$;

ALTER TABLE vampire_characters DROP COLUMN IF EXISTS post_act1_context;
