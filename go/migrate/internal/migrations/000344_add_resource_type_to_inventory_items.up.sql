ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS resource_type TEXT;

ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_resource_type_check;

ALTER TABLE inventory_items
  ADD CONSTRAINT inventory_items_resource_type_check
  CHECK (
    resource_type IS NULL OR resource_type IN (
      'herbalism',
      'mining',
      'logging',
      'skinning',
      'fishing'
    )
  );
