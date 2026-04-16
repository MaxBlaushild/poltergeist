ALTER TABLE resources
  ADD COLUMN IF NOT EXISTS inventory_item_id INTEGER;

ALTER TABLE resources
  DROP CONSTRAINT IF EXISTS resources_inventory_item_id_fkey;

ALTER TABLE resources
  ADD CONSTRAINT resources_inventory_item_id_fkey
  FOREIGN KEY (inventory_item_id)
  REFERENCES inventory_items(id)
  ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_resources_inventory_item_id
  ON resources(inventory_item_id);
