ALTER TABLE team_inventory_items
  DROP COLUMN inventory_item_id;
ALTER TABLE team_inventory_items
  ADD COLUMN inventory_item_id int;
CREATE INDEX idx_team_inventory_items_inventory_item_id ON team_inventory_items (inventory_item_id);

ALTER TABLE match_inventory_item_effects
  DROP COLUMN inventory_item_id;
ALTER TABLE match_inventory_item_effects
  ADD COLUMN inventory_item_id int;
CREATE INDEX idx_match_inventory_item_effects_inventory_item_id ON match_inventory_item_effects (inventory_item_id);
