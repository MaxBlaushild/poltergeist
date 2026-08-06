-- R-2.1: box internal depth is the master constraint everything else in the
-- set-assembly service checks against. depth_is_placeholder/measurement_notes
-- exist because the depth figure for a given box is often lower-confidence
-- than its length/width (see 000453's seed comment).
CREATE TABLE IF NOT EXISTS bgi_box_profiles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  game_id UUID NOT NULL REFERENCES bgi_games(id) ON DELETE CASCADE,
  slug TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'original' CHECK (source IN ('original', 'aftermarket')),
  interior_length_mm NUMERIC NOT NULL,
  interior_width_mm NUMERIC NOT NULL,
  interior_depth_mm NUMERIC NOT NULL,
  depth_is_placeholder BOOLEAN NOT NULL DEFAULT FALSE,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  source_url TEXT NOT NULL DEFAULT '',
  measurement_notes TEXT NOT NULL DEFAULT '',
  CONSTRAINT bgi_box_profiles_verified_requires_source CHECK (NOT verified OR source_url <> ''),
  UNIQUE (game_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_bgi_box_profiles_game_id ON bgi_box_profiles(game_id);
