CREATE TABLE user_equipment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equipment_slot VARCHAR(50) NOT NULL,
    owned_inventory_item_id UUID NOT NULL REFERENCES owned_inventory_items(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Ensure each user can only have one item equipped per slot
CREATE UNIQUE INDEX user_equipment_user_slot_unique ON user_equipment(user_id, equipment_slot);

-- Index for fast lookups by user
CREATE INDEX user_equipment_user_id_idx ON user_equipment(user_id);

-- Index for fast lookups by owned inventory item
CREATE INDEX user_equipment_owned_inventory_item_id_idx ON user_equipment(owned_inventory_item_id);