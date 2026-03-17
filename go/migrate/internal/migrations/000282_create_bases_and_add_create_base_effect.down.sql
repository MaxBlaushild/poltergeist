ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS consume_create_base;

DROP INDEX IF EXISTS idx_bases_geometry;
DROP INDEX IF EXISTS idx_bases_user_id;
DROP TABLE IF EXISTS bases;
