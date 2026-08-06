-- Structural mirror of reef_slice_results (000427), keyed by geometry_hash,
-- FK retargeted to bgi_tray_templates. sealed_void/min_wall_mm columns are
-- kept for schema symmetry with the shared validate.Metadata shape, but bgi
-- trays run with SealedVoidRuleEnabled=false (see go/pkg/reef/validate) —
-- open-top wells have no cavity story at all.
CREATE TABLE IF NOT EXISTS bgi_tray_slice_results (
  geometry_hash TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  tray_template_id UUID NOT NULL REFERENCES bgi_tray_templates(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('pending', 'valid', 'rejected')) DEFAULT 'pending',
  rejection_rule TEXT NOT NULL DEFAULT '',
  rejection_reason TEXT NOT NULL DEFAULT '',
  weight_g NUMERIC,
  print_time_s INTEGER,
  bbox_mm JSONB NOT NULL DEFAULT '{}'::jsonb,
  plate_fits BOOLEAN,
  support_required BOOLEAN,
  support_material_percent NUMERIC,
  min_wall_mm NUMERIC,
  sealed_void BOOLEAN,
  warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
  slicer_version TEXT NOT NULL DEFAULT '',
  openscad_version TEXT NOT NULL DEFAULT '',
  stl_key TEXT NOT NULL DEFAULT '',
  preview_key TEXT NOT NULL DEFAULT '',
  price_cents INTEGER
);

CREATE INDEX IF NOT EXISTS idx_bgi_tray_slice_results_tray_template_id ON bgi_tray_slice_results(tray_template_id);
CREATE INDEX IF NOT EXISTS idx_bgi_tray_slice_results_status ON bgi_tray_slice_results(status);
