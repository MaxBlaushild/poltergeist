CREATE TABLE IF NOT EXISTS reef_parameter_schemas (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES reef_products(id) ON DELETE CASCADE,
  version INTEGER NOT NULL,
  schema JSONB NOT NULL,
  generator_module TEXT NOT NULL,
  generator_version TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (product_id, version)
);

CREATE INDEX IF NOT EXISTS idx_reef_parameter_schemas_product_active
  ON reef_parameter_schemas(product_id, active);
