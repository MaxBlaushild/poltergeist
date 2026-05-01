ALTER TABLE zone_kinds
ADD COLUMN IF NOT EXISTS default_shopkeeper_item_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
