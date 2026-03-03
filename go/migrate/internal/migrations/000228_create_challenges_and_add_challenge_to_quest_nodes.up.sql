DO $$
BEGIN
  -- Legacy schema from 000014 had user_id/challenge columns.
  -- Preserve it under a backup name before creating the new map challenge table.
  IF to_regclass('public.challenges') IS NOT NULL
    AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'challenges'
        AND column_name = 'user_id'
    )
    AND NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'challenges'
        AND column_name = 'zone_id'
    )
    AND to_regclass('public.challenges_legacy_000228') IS NULL THEN
    ALTER TABLE challenges RENAME TO challenges_legacy_000228;
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS challenges (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326),
  question TEXT NOT NULL,
  reward INTEGER NOT NULL DEFAULT 0,
  inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE SET NULL,
  submission_type TEXT NOT NULL DEFAULT 'photo',
  difficulty INTEGER NOT NULL DEFAULT 0,
  stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency TEXT
);

CREATE INDEX IF NOT EXISTS idx_challenges_zone_id ON challenges(zone_id);
CREATE INDEX IF NOT EXISTS idx_challenges_geometry ON challenges USING GIST(geometry);

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS challenge_id UUID REFERENCES challenges(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS quest_nodes_challenge_idx ON quest_nodes(challenge_id);
