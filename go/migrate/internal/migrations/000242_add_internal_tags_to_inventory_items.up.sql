ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS internal_tags JSONB;

UPDATE inventory_items
SET internal_tags = '[]'::jsonb
WHERE internal_tags IS NULL;

ALTER TABLE inventory_items
  ALTER COLUMN internal_tags SET DEFAULT '[]'::jsonb,
  ALTER COLUMN internal_tags SET NOT NULL;
