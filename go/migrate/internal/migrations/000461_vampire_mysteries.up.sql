-- Mysteries: the underlying story an instance's players are solving. See
-- go/vampire-ascendancy/docs/MYSTERY_REQUIREMENTS.md.
CREATE TABLE IF NOT EXISTS vampire_mysteries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  full_lore TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Beats: an ordered list of discoverable facts about the mystery. Secrets
-- point at the beat they reveal (many secrets may point at the same beat).
CREATE TABLE IF NOT EXISTS vampire_mystery_beats (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  mystery_id UUID NOT NULL REFERENCES vampire_mysteries(id) ON DELETE CASCADE,
  ordinal INT NOT NULL DEFAULT 0,
  body TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_vampire_mystery_beats_mystery_id ON vampire_mystery_beats(mystery_id);

-- Secrets become (character, mystery) scoped instead of just per-character,
-- and point at the beat they reveal. Nullable indefinitely (not just during
-- backfill) — a secret with no mystery simply isn't usable by any instance,
-- and beat_id is a form-level required field in the Super Admin editor, not
-- a hard DB constraint.
ALTER TABLE vampire_secrets ADD COLUMN IF NOT EXISTS mystery_id UUID REFERENCES vampire_mysteries(id);
ALTER TABLE vampire_secrets ADD COLUMN IF NOT EXISTS beat_id UUID REFERENCES vampire_mystery_beats(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_vampire_secrets_mystery_id ON vampire_secrets(mystery_id);
CREATE INDEX IF NOT EXISTS idx_vampire_secrets_character_mystery ON vampire_secrets(character_id, mystery_id);

-- Quiz questions belong to exactly one mystery. Nullable now, NOT NULL in a
-- follow-up migration once the backfill below is confirmed.
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS mystery_id UUID REFERENCES vampire_mysteries(id);
CREATE INDEX IF NOT EXISTS idx_vampire_quiz_questions_mystery_id ON vampire_quiz_questions(mystery_id);

-- An instance is one mystery, for its whole life (chosen at "Host a Toast"
-- time, never changed after). Nullable now, NOT NULL in a follow-up
-- migration once the backfill below is confirmed.
ALTER TABLE vampire_instances ADD COLUMN IF NOT EXISTS mystery_id UUID REFERENCES vampire_mysteries(id);

-- Backfill: one mystery row for the live story, and point every existing
-- quiz question / secret / instance at it.
DO $$
DECLARE
  legacy_mystery_id UUID;
BEGIN
  SELECT id INTO legacy_mystery_id FROM vampire_mysteries WHERE name = 'The Crimson Toast' LIMIT 1;
  IF legacy_mystery_id IS NULL THEN
    INSERT INTO vampire_mysteries (name, summary, full_lore)
    VALUES (
      'The Crimson Toast',
      'A murder at the annual Crimson Toast unravels into betrayal, forbidden bloodlines, and a fight over who inherits the Court.',
      ''
    )
    RETURNING id INTO legacy_mystery_id;
  END IF;

  UPDATE vampire_quiz_questions SET mystery_id = legacy_mystery_id WHERE mystery_id IS NULL;
  UPDATE vampire_secrets SET mystery_id = legacy_mystery_id WHERE mystery_id IS NULL;
  UPDATE vampire_instances SET mystery_id = legacy_mystery_id WHERE mystery_id IS NULL;
END $$;

-- The backfill above runs in the same migration/transaction, so there's no
-- real "confirm, then tighten" checkpoint in this deploy workflow to wait
-- on — every pending migration applies in one batch. Safe to require going
-- forward immediately: quiz_questions/instances always get their mystery_id
-- through server-side code paths that now set it (adminCreateQuizQuestion,
-- createInstance), never left null for new rows.
ALTER TABLE vampire_quiz_questions ALTER COLUMN mystery_id SET NOT NULL;
ALTER TABLE vampire_instances ALTER COLUMN mystery_id SET NOT NULL;

-- vampire_instance_quiz_questions is fully superseded by mystery-scoping —
-- an instance's quiz is now just "questions where mystery_id matches",
-- not a per-instance toggle. Drop it; the down migration recreates the
-- empty table (schema only, no data) rather than block a revert.
DROP TABLE IF EXISTS vampire_instance_quiz_questions;
