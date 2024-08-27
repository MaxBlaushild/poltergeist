ALTER TABLE match_inventory_items
ADD COLUMN team_id UUID;

ALTER TABLE match_inventory_items
ADD CONSTRAINT fk_team_id
FOREIGN KEY (team_id) REFERENCES teams(id);
