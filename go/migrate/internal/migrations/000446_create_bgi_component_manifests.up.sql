-- R-2.4: component counts/dimensions are facts, not designs, and are the
-- backbone of both tray design and fit validation. NULL expansion_id means
-- the row belongs to the base game.
CREATE TABLE IF NOT EXISTS bgi_component_manifests (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  game_id UUID NOT NULL REFERENCES bgi_games(id) ON DELETE CASCADE,
  expansion_id UUID REFERENCES bgi_expansions(id) ON DELETE CASCADE,
  component_type TEXT NOT NULL,
  card_width_mm NUMERIC,
  card_height_mm NUMERIC,
  count INTEGER NOT NULL,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  source_url TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  CONSTRAINT bgi_component_manifests_verified_requires_source CHECK (NOT verified OR source_url <> ''),
  UNIQUE (game_id, expansion_id, component_type)
);

CREATE INDEX IF NOT EXISTS idx_bgi_component_manifests_game_id ON bgi_component_manifests(game_id);
