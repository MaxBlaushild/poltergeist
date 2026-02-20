DROP INDEX IF EXISTS idx_zones_zone_import_id;
ALTER TABLE zones DROP COLUMN IF EXISTS zone_import_id;
