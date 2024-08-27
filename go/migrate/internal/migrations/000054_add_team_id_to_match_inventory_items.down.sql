ALTER TABLE match_inventory_items
DROP CONSTRAINT fk_team_id;

ALTER TABLE match_inventory_items
DROP COLUMN team_id;
