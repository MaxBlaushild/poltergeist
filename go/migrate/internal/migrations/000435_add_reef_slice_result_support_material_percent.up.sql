-- R-5.2's supports rule moved from a bare boolean ("any support at all
-- rejects") to a percentage threshold (go/pkg/reef/validate's
-- checkExcessiveSupport), computed from real extrusion-length tracking in
-- the sliced G-code (go/pkg/reef/slice.ParseGCodeStats). support_required
-- stays as a derived/display convenience (percent > 0).
ALTER TABLE reef_slice_results
  ADD COLUMN IF NOT EXISTS support_material_percent DOUBLE PRECISION;
