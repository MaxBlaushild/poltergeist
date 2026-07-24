-- Keyed by geometry_hash so identical parameters across configurations, sessions,
-- and time never regenerate or re-slice (R-3.3). One row per unique geometry.
CREATE TABLE IF NOT EXISTS reef_slice_results (
  geometry_hash TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  product_id UUID NOT NULL REFERENCES reef_products(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('pending', 'valid', 'rejected')) DEFAULT 'pending',
  rejection_rule TEXT NOT NULL DEFAULT '',
  rejection_reason TEXT NOT NULL DEFAULT '',
  weight_g NUMERIC,
  print_time_s INTEGER,
  bbox_mm JSONB NOT NULL DEFAULT '{}'::jsonb,
  plate_fits BOOLEAN,
  support_required BOOLEAN,
  min_wall_mm NUMERIC,
  sealed_void BOOLEAN,
  warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
  slicer_version TEXT NOT NULL DEFAULT '',
  openscad_version TEXT NOT NULL DEFAULT '',
  stl_key TEXT NOT NULL DEFAULT '',
  preview_key TEXT NOT NULL DEFAULT '',
  price_cents INTEGER
);

CREATE INDEX IF NOT EXISTS idx_reef_slice_results_product_id ON reef_slice_results(product_id);
CREATE INDEX IF NOT EXISTS idx_reef_slice_results_status ON reef_slice_results(status);
