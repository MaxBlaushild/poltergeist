CREATE TABLE IF NOT EXISTS reef_configurations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES reef_products(id) ON DELETE CASCADE,
  params JSONB NOT NULL,
  geometry_hash TEXT REFERENCES reef_slice_results(geometry_hash),
  status TEXT NOT NULL CHECK (status IN ('pending', 'valid', 'rejected')) DEFAULT 'pending',
  rejection_reason TEXT NOT NULL DEFAULT '',
  price_cents INTEGER,
  session_id TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_reef_configurations_product_id ON reef_configurations(product_id);
CREATE INDEX IF NOT EXISTS idx_reef_configurations_geometry_hash ON reef_configurations(geometry_hash);
CREATE INDEX IF NOT EXISTS idx_reef_configurations_params_gin ON reef_configurations USING GIN (params);

CREATE TABLE IF NOT EXISTS reef_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  configuration_id UUID NOT NULL REFERENCES reef_configurations(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('preview', 'full')),
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed')) DEFAULT 'queued',
  attempts INTEGER NOT NULL DEFAULT 0,
  locked_at TIMESTAMPTZ,
  error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_reef_generation_jobs_configuration_id ON reef_generation_jobs(configuration_id);
CREATE INDEX IF NOT EXISTS idx_reef_generation_jobs_status ON reef_generation_jobs(status);
