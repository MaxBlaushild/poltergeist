DROP INDEX IF EXISTS user_equipment_owned_item_idx;
CREATE UNIQUE INDEX user_equipment_owned_item_idx ON user_equipment(owned_inventory_item_id);
