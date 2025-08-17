CREATE TABLE user_equipment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    helm_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    chest_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    left_hand_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    right_hand_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    feet_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    gloves_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    neck_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    left_ring_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    right_ring_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    leg_inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX user_equipment_user_id_idx ON user_equipment(user_id);