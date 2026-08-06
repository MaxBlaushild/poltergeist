-- Mirrors reef_products/reef_parameter_schemas (000424/000425) exactly. A
-- bgi product's generator_module on its active schema is the sentinel
-- 'bgi_tray_set' — go/bgi-site's configure handlers recognize this and route
-- through go/pkg/reef/set.Assemble instead of generate.Get(), since a tray
-- set is many generated parts, not one (R-3.3).
CREATE TABLE IF NOT EXISTS bgi_products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  game_id UUID NOT NULL REFERENCES bgi_games(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('configurable', 'fixed')),
  description TEXT NOT NULL DEFAULT '',
  material TEXT NOT NULL DEFAULT 'PETG',
  base_price_cents INTEGER NOT NULL DEFAULT 0,
  images JSONB NOT NULL DEFAULT '[]'::jsonb,
  active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_bgi_products_active ON bgi_products(active);
CREATE INDEX IF NOT EXISTS idx_bgi_products_game_id ON bgi_products(game_id);

CREATE TABLE IF NOT EXISTS bgi_parameter_schemas (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES bgi_products(id) ON DELETE CASCADE,
  version INTEGER NOT NULL,
  schema JSONB NOT NULL,
  generator_module TEXT NOT NULL,
  generator_version TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (product_id, version)
);

CREATE INDEX IF NOT EXISTS idx_bgi_parameter_schemas_product_active
  ON bgi_parameter_schemas(product_id, active);
