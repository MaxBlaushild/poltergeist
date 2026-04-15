ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS resource_type TEXT;

UPDATE inventory_items AS ii
SET resource_type = rt.slug
FROM resource_types AS rt
WHERE ii.resource_type_id = rt.id
  AND (ii.resource_type IS NULL OR TRIM(ii.resource_type) = '');

ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_resource_type_id_fkey;

DROP INDEX IF EXISTS idx_inventory_items_resource_type_id;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS resource_type_id;

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

DROP TABLE IF EXISTS user_resource_gatherings;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS resource_types;
