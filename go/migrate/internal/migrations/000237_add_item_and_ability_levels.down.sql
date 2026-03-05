ALTER TABLE spells
  DROP COLUMN IF EXISTS ability_level;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS item_level;
