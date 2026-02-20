ALTER TABLE zones
  ADD COLUMN IF NOT EXISTS zone_import_id UUID REFERENCES zone_imports(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_zones_zone_import_id ON zones(zone_import_id);
