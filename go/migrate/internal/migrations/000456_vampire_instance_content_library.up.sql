-- Vampire Ascendancy multi-tenancy, part 3: the content library stays
-- global and shared (vampire_characters/items/quiz_questions), but which of
-- it is included in a given instance, plus each character's sigil and
-- portrait, are inherently per-instance — two concurrent instances must be
-- able to give "Valen Drear" different PINs and different player-submitted
-- photos. See MULTI_TENANT_REQUIREMENTS.md.

CREATE TABLE IF NOT EXISTS vampire_instance_characters (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  included BOOLEAN NOT NULL DEFAULT TRUE,
  -- Moved off vampire_characters; see the deprecation note below.
  sigil TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, character_id)
);
CREATE INDEX IF NOT EXISTS vampire_instance_characters_instance_idx ON vampire_instance_characters(instance_id);

CREATE TABLE IF NOT EXISTS vampire_instance_items (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  item_id UUID NOT NULL REFERENCES vampire_items(id) ON DELETE CASCADE,
  included BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, item_id)
);
CREATE INDEX IF NOT EXISTS vampire_instance_items_instance_idx ON vampire_instance_items(instance_id);

CREATE TABLE IF NOT EXISTS vampire_instance_quiz_questions (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES vampire_quiz_questions(id) ON DELETE CASCADE,
  included BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, question_id)
);
CREATE INDEX IF NOT EXISTS vampire_instance_quiz_questions_instance_idx ON vampire_instance_quiz_questions(instance_id);

-- Backfill: the legacy instance (created in 000455) includes everything
-- currently in the library, carrying over each character's existing sigil
-- and portrait. Looked up inline (not cached in a temp table) — temp
-- tables live for the whole migration session, not just one file, so a
-- name reused across migrations (as 000455 does) would collide.
INSERT INTO vampire_instance_characters (instance_id, character_id, included, sigil, image_url)
  SELECT (SELECT id FROM vampire_instances WHERE name = 'The Crimson Toast' ORDER BY created_at ASC LIMIT 1), c.id, TRUE, c.password, c.image_url
  FROM vampire_characters c
  ON CONFLICT (instance_id, character_id) DO NOTHING;

INSERT INTO vampire_instance_items (instance_id, item_id, included)
  SELECT (SELECT id FROM vampire_instances WHERE name = 'The Crimson Toast' ORDER BY created_at ASC LIMIT 1), i.id, TRUE
  FROM vampire_items i
  ON CONFLICT (instance_id, item_id) DO NOTHING;

INSERT INTO vampire_instance_quiz_questions (instance_id, question_id, included)
  SELECT (SELECT id FROM vampire_instances WHERE name = 'The Crimson Toast' ORDER BY created_at ASC LIMIT 1), q.id, TRUE
  FROM vampire_quiz_questions q
  ON CONFLICT (instance_id, question_id) DO NOTHING;

-- Deprecated in place: sigil/portrait now live on vampire_instance_characters.
-- Kept on vampire_characters for one release as a safety net, then dropped
-- in a follow-up migration once the app is verified to read/write the new
-- columns exclusively.
COMMENT ON COLUMN vampire_characters.password IS 'DEPRECATED — moved to vampire_instance_characters.sigil. Do not read/write.';
COMMENT ON COLUMN vampire_characters.image_url IS 'DEPRECATED — moved to vampire_instance_characters.image_url. Do not read/write.';
