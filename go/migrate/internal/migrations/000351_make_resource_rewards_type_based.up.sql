ALTER TABLE resources
  DROP CONSTRAINT IF EXISTS resources_inventory_item_id_fkey;

DROP INDEX IF EXISTS idx_resources_inventory_item_id;

ALTER TABLE resources
  DROP COLUMN IF EXISTS inventory_item_id;
