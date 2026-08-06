-- bgi_set_resolutions is the config_hash cache (R-4.3): identical
-- (game, expansions, sleeve, box, color) combinations resolve to one
-- assembled recipe. Individual trays still cache independently by their own
-- geometry_hash in bgi_tray_slice_results, so two different config_hashes
-- can share the same underlying tray geometry.
CREATE TABLE IF NOT EXISTS bgi_set_resolutions (
  config_hash TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES bgi_products(id) ON DELETE CASCADE,
  game_id UUID NOT NULL REFERENCES bgi_games(id),
  expansion_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  sleeve_profile_id UUID NOT NULL REFERENCES bgi_sleeve_profiles(id),
  box_profile_id UUID NOT NULL REFERENCES bgi_box_profiles(id),
  resolved_trays JSONB NOT NULL,
  unassembled_components JSONB NOT NULL DEFAULT '[]'::jsonb,
  assembled_height_mm NUMERIC,
  fits_box BOOLEAN
);

CREATE INDEX IF NOT EXISTS idx_bgi_set_resolutions_product_id ON bgi_set_resolutions(product_id);

CREATE TABLE IF NOT EXISTS bgi_configurations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES bgi_products(id) ON DELETE CASCADE,
  params JSONB NOT NULL,
  config_hash TEXT REFERENCES bgi_set_resolutions(config_hash),
  status TEXT NOT NULL CHECK (status IN ('pending', 'valid', 'rejected')) DEFAULT 'pending',
  rejection_reason TEXT NOT NULL DEFAULT '',
  price_cents INTEGER,
  session_id TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_bgi_configurations_product_id ON bgi_configurations(product_id);
CREATE INDEX IF NOT EXISTS idx_bgi_configurations_config_hash ON bgi_configurations(config_hash);
CREATE INDEX IF NOT EXISTS idx_bgi_configurations_params_gin ON bgi_configurations USING GIN (params);

CREATE TABLE IF NOT EXISTS bgi_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  configuration_id UUID NOT NULL REFERENCES bgi_configurations(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('full_set')),
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed')) DEFAULT 'queued',
  attempts INTEGER NOT NULL DEFAULT 0,
  locked_at TIMESTAMPTZ,
  error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_bgi_generation_jobs_configuration_id ON bgi_generation_jobs(configuration_id);
CREATE INDEX IF NOT EXISTS idx_bgi_generation_jobs_status ON bgi_generation_jobs(status);
