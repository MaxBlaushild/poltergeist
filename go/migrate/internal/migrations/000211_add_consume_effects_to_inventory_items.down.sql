ALTER TABLE inventory_items
DROP COLUMN IF EXISTS consume_statuses_to_remove,
DROP COLUMN IF EXISTS consume_statuses_to_add,
DROP COLUMN IF EXISTS consume_mana_delta,
DROP COLUMN IF EXISTS consume_health_delta;
