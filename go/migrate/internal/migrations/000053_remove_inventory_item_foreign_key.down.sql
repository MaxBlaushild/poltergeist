DROP INDEX IF EXISTS idx_match_inventory_item_effects_inventory_item_id;
ALTER TABLE match_inventory_item_effects
  ALTER COLUMN inventory_item_id TYPE varchar USING (inventory_item_id::varchar);

DROP INDEX IF EXISTS idx_team_inventory_items_inventory_item_id;
ALTER TABLE team_inventory_items
  ALTER COLUMN inventory_item_id TYPE varchar USING (inventory_item_id::varchar);
