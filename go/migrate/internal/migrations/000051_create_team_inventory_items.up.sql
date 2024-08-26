CREATE TABLE team_inventory_items (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL,
    inventory_item_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id),
    FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id)
);
