DROP INDEX IF EXISTS idx_monsters_off_hand_inventory_item_id;
DROP INDEX IF EXISTS idx_monsters_dominant_hand_inventory_item_id;

ALTER TABLE monsters
  DROP CONSTRAINT IF EXISTS fk_monsters_off_hand_inventory_item_id;

ALTER TABLE monsters
  DROP CONSTRAINT IF EXISTS fk_monsters_dominant_hand_inventory_item_id;

ALTER TABLE monsters
  DROP COLUMN IF EXISTS off_hand_inventory_item_id,
  DROP COLUMN IF EXISTS dominant_hand_inventory_item_id;
