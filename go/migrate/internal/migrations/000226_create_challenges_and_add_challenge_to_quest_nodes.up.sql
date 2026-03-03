CREATE TABLE challenges (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326),
  question TEXT NOT NULL,
  reward INTEGER NOT NULL DEFAULT 0,
  inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE SET NULL,
  submission_type TEXT NOT NULL DEFAULT 'photo',
  difficulty INTEGER NOT NULL DEFAULT 0,
  stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency TEXT
);

CREATE INDEX idx_challenges_zone_id ON challenges(zone_id);
CREATE INDEX idx_challenges_geometry ON challenges USING GIST(geometry);

ALTER TABLE quest_nodes
  ADD COLUMN challenge_id UUID REFERENCES challenges(id) ON DELETE SET NULL;

CREATE INDEX quest_nodes_challenge_idx ON quest_nodes(challenge_id);
