-- Drop the unique index
DROP INDEX IF EXISTS idx_inventory_items_inventory_item_id;

-- Remove the new columns
ALTER TABLE inventory_items 
DROP COLUMN IF EXISTS inventory_item_id,
DROP COLUMN IF EXISTS rarity_tier,
DROP COLUMN IF EXISTS is_capture_type,
DROP COLUMN IF EXISTS item_type,
DROP COLUMN IF EXISTS equipment_slot;

-- Clear the seeded data
DELETE FROM inventory_items;