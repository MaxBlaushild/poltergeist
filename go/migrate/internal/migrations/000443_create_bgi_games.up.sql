-- bgi-site (board game insert organizers), the second microsite on the
-- shared reef print-and-slice platform (go/pkg/reef/*). See
-- go/bgi-site/PLATFORM_FINDINGS.md for what generalized from reef's build
-- and what didn't.
CREATE TABLE IF NOT EXISTS bgi_games (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  publisher TEXT NOT NULL DEFAULT '',
  year_published INTEGER,
  -- Gates catalog visibility (same meaning as reef_products.active), not a
  -- data-trust signal — see bgi_box_profiles/bgi_sleeve_profiles/
  -- bgi_component_manifests' own verified columns for that; the frontend
  -- must surface a "pending physical verification" notice wherever that
  -- data is used rather than hiding the whole game behind this flag.
  active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_bgi_games_active ON bgi_games(active);

CREATE TABLE IF NOT EXISTS bgi_expansions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  game_id UUID NOT NULL REFERENCES bgi_games(id) ON DELETE CASCADE,
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  standalone BOOLEAN NOT NULL DEFAULT FALSE,
  active BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE (game_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_bgi_expansions_game_id ON bgi_expansions(game_id);
