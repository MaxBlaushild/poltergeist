ALTER TABLE monsters
  ADD COLUMN IF NOT EXISTS dominant_hand_inventory_item_id INTEGER,
  ADD COLUMN IF NOT EXISTS off_hand_inventory_item_id INTEGER;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_monsters_dominant_hand_inventory_item_id'
  ) THEN
    ALTER TABLE monsters
      ADD CONSTRAINT fk_monsters_dominant_hand_inventory_item_id
      FOREIGN KEY (dominant_hand_inventory_item_id) REFERENCES inventory_items(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_monsters_off_hand_inventory_item_id'
  ) THEN
    ALTER TABLE monsters
      ADD CONSTRAINT fk_monsters_off_hand_inventory_item_id
      FOREIGN KEY (off_hand_inventory_item_id) REFERENCES inventory_items(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_monsters_dominant_hand_inventory_item_id ON monsters(dominant_hand_inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_monsters_off_hand_inventory_item_id ON monsters(off_hand_inventory_item_id);

UPDATE monsters
SET dominant_hand_inventory_item_id = weapon_inventory_item_id
WHERE dominant_hand_inventory_item_id IS NULL AND weapon_inventory_item_id IS NOT NULL;
