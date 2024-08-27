ALTER TABLE match_inventory_item_effects
DROP CONSTRAINT fk_team_id;

ALTER TABLE match_inventory_item_effects
DROP COLUMN team_id;
