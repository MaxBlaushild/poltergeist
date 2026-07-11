-- One GM-facing reference photo per catalog item (so a GM can see which physical
-- prop maps to a dashboard item). Bytes stored inline, like submission photos.
CREATE TABLE IF NOT EXISTS vampire_item_photos (
    item_id UUID PRIMARY KEY REFERENCES vampire_items(id) ON DELETE CASCADE,
    content_type TEXT NOT NULL,
    data BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
