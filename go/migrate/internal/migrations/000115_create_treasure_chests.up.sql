CREATE TABLE treasure_chests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    zone_id UUID NOT NULL REFERENCES zones(id),
    gold INTEGER,
    geometry GEOMETRY(Point, 4326)
);

CREATE INDEX idx_treasure_chests_zone_id ON treasure_chests(zone_id);
CREATE INDEX idx_treasure_chests_geometry ON treasure_chests USING GIST(geometry);

CREATE TABLE treasure_chest_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    treasure_chest_id UUID NOT NULL REFERENCES treasure_chests(id) ON DELETE CASCADE,
    inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    UNIQUE(treasure_chest_id, inventory_item_id)
);

CREATE INDEX idx_treasure_chest_items_treasure_chest_id ON treasure_chest_items(treasure_chest_id);
CREATE INDEX idx_treasure_chest_items_inventory_item_id ON treasure_chest_items(inventory_item_id);

