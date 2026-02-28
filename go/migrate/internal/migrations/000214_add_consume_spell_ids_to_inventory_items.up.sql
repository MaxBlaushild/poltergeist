ALTER TABLE inventory_items
ADD COLUMN consume_spell_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
