ALTER TABLE inventory_items
ADD COLUMN buy_price INTEGER;

UPDATE inventory_items
SET buy_price = sell_value * 2
WHERE sell_value IS NOT NULL;

ALTER TABLE inventory_items
DROP COLUMN IF EXISTS sell_value;
