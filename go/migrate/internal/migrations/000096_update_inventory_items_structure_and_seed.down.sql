-- Remove all seeded data
DELETE FROM inventory_items;

-- Remove the new columns
ALTER TABLE inventory_items DROP COLUMN equipment_slot;
ALTER TABLE inventory_items DROP COLUMN item_type;
ALTER TABLE inventory_items DROP COLUMN is_capture_type;
ALTER TABLE inventory_items DROP COLUMN rarity_tier;

-- Change id back to UUID
ALTER TABLE inventory_items DROP CONSTRAINT inventory_items_pkey;
ALTER TABLE inventory_items DROP COLUMN id;
ALTER TABLE inventory_items ADD COLUMN id UUID PRIMARY KEY;