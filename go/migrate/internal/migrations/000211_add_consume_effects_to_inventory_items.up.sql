ALTER TABLE inventory_items
ADD COLUMN consume_health_delta INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consume_mana_delta INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consume_statuses_to_add JSONB NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN consume_statuses_to_remove JSONB NOT NULL DEFAULT '[]'::jsonb;
