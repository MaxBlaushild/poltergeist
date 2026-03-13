ALTER TABLE inventory_items
ADD COLUMN sell_value INTEGER;

UPDATE inventory_items
SET sell_value = buy_price / 2
WHERE buy_price IS NOT NULL;

ALTER TABLE inventory_items
DROP COLUMN IF EXISTS buy_price;
