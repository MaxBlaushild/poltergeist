CREATE TABLE match_inventory_item_effects (
    id UUID PRIMARY KEY,
    match_id UUID NOT NULL,
    inventory_item_id UUID NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id)
);

CREATE INDEX idx_match_inventory_item_effects_expires_at ON match_inventory_item_effects(expires_at);
