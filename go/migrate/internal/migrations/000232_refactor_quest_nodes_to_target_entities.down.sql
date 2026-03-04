ALTER TABLE quest_nodes
  DROP CONSTRAINT IF EXISTS quest_nodes_exactly_one_target;

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS point_of_interest_id UUID REFERENCES points_of_interest(id),
  ADD COLUMN IF NOT EXISTS polygon geometry(Polygon,4326);

CREATE INDEX IF NOT EXISTS quest_nodes_poi_idx ON quest_nodes(point_of_interest_id);
