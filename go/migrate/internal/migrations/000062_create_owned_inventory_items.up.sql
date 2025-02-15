CREATE TABLE owned_inventory_items (
    id UUID PRIMARY KEY,
    team_id UUID REFERENCES teams(id),
    user_id UUID REFERENCES users(id),
    inventory_item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX owned_inventory_items_team_id_idx ON owned_inventory_items(team_id);
CREATE INDEX owned_inventory_items_user_id_idx ON owned_inventory_items(user_id);
CREATE INDEX owned_inventory_items_inventory_item_id_idx ON owned_inventory_items(inventory_item_id);
