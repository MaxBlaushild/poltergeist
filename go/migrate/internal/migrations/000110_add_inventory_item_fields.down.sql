ALTER TABLE inventory_items
DROP COLUMN IF EXISTS rarity_tier,
DROP COLUMN IF EXISTS is_capture_type;

