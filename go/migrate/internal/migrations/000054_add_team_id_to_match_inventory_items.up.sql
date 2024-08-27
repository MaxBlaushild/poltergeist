ALTER TABLE match_inventory_item_effects
ADD COLUMN team_id UUID;

ALTER TABLE match_inventory_item_effects
ADD CONSTRAINT fk_team_id
FOREIGN KEY (team_id) REFERENCES teams(id);
