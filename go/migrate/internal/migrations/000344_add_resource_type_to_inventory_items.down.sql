ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_resource_type_check;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS resource_type;
