ALTER TABLE monster_templates
DROP COLUMN IF EXISTS archived;

ALTER TABLE inventory_items
DROP COLUMN IF EXISTS archived;
