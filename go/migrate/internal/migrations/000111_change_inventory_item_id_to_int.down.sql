-- Drop foreign key constraints
ALTER TABLE match_inventory_item_effects DROP CONSTRAINT IF EXISTS match_inventory_item_effects_inventory_item_id_fkey;
ALTER TABLE team_inventory_items DROP CONSTRAINT IF EXISTS team_inventory_items_inventory_item_id_fkey;

-- Create a new table with UUID id
CREATE TABLE inventory_items_new (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    image_url TEXT,
    flavor_text TEXT,
    effect_text TEXT,
    rarity_tier TEXT,
    is_capture_type BOOLEAN DEFAULT FALSE
);

-- Copy existing data (note: we lose the original UUID mapping)
INSERT INTO inventory_items_new (name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, created_at, updated_at)
SELECT name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, created_at, updated_at
FROM inventory_items;

-- Drop the old table
DROP TABLE inventory_items;

-- Rename the new table
ALTER TABLE inventory_items_new RENAME TO inventory_items;

