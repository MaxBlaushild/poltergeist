-- First, drop any foreign key constraints that reference inventory_items.id (if they exist)
ALTER TABLE match_inventory_item_effects DROP CONSTRAINT IF EXISTS match_inventory_item_effects_inventory_item_id_fkey;
ALTER TABLE team_inventory_items DROP CONSTRAINT IF EXISTS team_inventory_items_inventory_item_id_fkey;

-- Create a new table with INTEGER id
CREATE TABLE inventory_items_new (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    image_url TEXT,
    flavor_text TEXT,
    effect_text TEXT,
    rarity_tier TEXT,
    is_capture_type BOOLEAN DEFAULT FALSE
);

-- Copy any existing data (if any) - handle the case where id might be UUID
-- We'll use a sequence to assign new IDs
INSERT INTO inventory_items_new (name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, created_at, updated_at)
SELECT name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, created_at, updated_at
FROM inventory_items
ORDER BY created_at;

-- Drop the old table
DROP TABLE inventory_items CASCADE;

-- Rename the new table
ALTER TABLE inventory_items_new RENAME TO inventory_items;

-- Rename the sequence to match the new table name
ALTER SEQUENCE inventory_items_new_id_seq RENAME TO inventory_items_id_seq;

-- Recreate indexes if needed
CREATE INDEX IF NOT EXISTS idx_inventory_items_id ON inventory_items(id);

