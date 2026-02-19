CREATE TABLE outfit_profile_generations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    owned_inventory_item_id UUID NOT NULL REFERENCES owned_inventory_items(id) ON DELETE CASCADE,
    inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    outfit_name TEXT NOT NULL,
    selfie_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    error_message TEXT,
    profile_picture_url TEXT
);

CREATE INDEX outfit_profile_generations_user_id_idx ON outfit_profile_generations(user_id);
CREATE INDEX outfit_profile_generations_owned_item_id_idx ON outfit_profile_generations(owned_inventory_item_id);
