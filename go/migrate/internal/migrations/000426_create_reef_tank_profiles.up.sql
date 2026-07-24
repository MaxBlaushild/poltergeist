CREATE TABLE IF NOT EXISTS reef_tank_profiles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  manufacturer TEXT NOT NULL,
  model TEXT NOT NULL,
  rim_thickness_mm NUMERIC NOT NULL,
  rim_width_mm NUMERIC NOT NULL,
  glass_thickness_mm NUMERIC NOT NULL,
  euro_brace BOOLEAN NOT NULL DEFAULT FALSE,
  internal_dims JSONB NOT NULL DEFAULT '{}'::jsonb,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  source_url TEXT NOT NULL DEFAULT '',
  UNIQUE (manufacturer, model),
  -- R-3.4: a row can only be marked verified once it carries a real source_url.
  CONSTRAINT reef_tank_profiles_verified_requires_source
    CHECK (NOT verified OR source_url <> '')
);

CREATE INDEX IF NOT EXISTS idx_reef_tank_profiles_verified ON reef_tank_profiles(verified);
