CREATE TABLE inventory_items (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    name TEXT NOT NULL,
    image_url TEXT,
    flavor_text TEXT,
    effect_text TEXT
);
