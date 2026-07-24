CREATE TABLE IF NOT EXISTS reef_products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('configurable', 'fixed')),
  description TEXT NOT NULL DEFAULT '',
  material TEXT NOT NULL DEFAULT 'PETG',
  base_price_cents INTEGER NOT NULL DEFAULT 0,
  images JSONB NOT NULL DEFAULT '[]'::jsonb,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_reef_products_active ON reef_products(active);

CREATE TABLE IF NOT EXISTS reef_product_variants (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES reef_products(id) ON DELETE CASCADE,
  variant_key TEXT NOT NULL,
  label TEXT NOT NULL,
  price_cents INTEGER NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (product_id, variant_key)
);

CREATE INDEX IF NOT EXISTS idx_reef_product_variants_product_id ON reef_product_variants(product_id);
