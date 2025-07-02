CREATE TABLE inventory_item_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    inventory_item_id INTEGER NOT NULL,
    strength_bonus INTEGER NOT NULL DEFAULT 0,
    dexterity_bonus INTEGER NOT NULL DEFAULT 0,
    constitution_bonus INTEGER NOT NULL DEFAULT 0,
    intelligence_bonus INTEGER NOT NULL DEFAULT 0,
    wisdom_bonus INTEGER NOT NULL DEFAULT 0,
    charisma_bonus INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id) ON DELETE CASCADE,
    UNIQUE(inventory_item_id)
);

CREATE INDEX inventory_item_stats_inventory_item_id_idx ON inventory_item_stats(inventory_item_id);