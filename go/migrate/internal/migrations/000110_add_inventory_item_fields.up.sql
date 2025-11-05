ALTER TABLE inventory_items
ADD COLUMN rarity_tier TEXT,
ADD COLUMN is_capture_type BOOLEAN DEFAULT FALSE;

