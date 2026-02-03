CREATE TABLE quests (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  image_url TEXT,
  zone_id UUID REFERENCES zones(id),
  quest_archetype_id UUID REFERENCES quest_archetypes(id),
  quest_giver_character_id UUID REFERENCES characters(id),
  gold INTEGER DEFAULT 0
);

CREATE TABLE quest_nodes (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  order_index INTEGER NOT NULL,
  point_of_interest_id UUID REFERENCES points_of_interest(id),
  polygon geometry(Polygon,4326)
);

CREATE INDEX quest_nodes_quest_id_idx ON quest_nodes(quest_id);
CREATE INDEX quest_nodes_poi_idx ON quest_nodes(point_of_interest_id);

CREATE TABLE quest_node_challenges (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_node_id UUID NOT NULL REFERENCES quest_nodes(id) ON DELETE CASCADE,
  tier INTEGER NOT NULL,
  question TEXT NOT NULL,
  reward INTEGER NOT NULL,
  inventory_item_id INTEGER
);

CREATE INDEX quest_node_challenges_node_idx ON quest_node_challenges(quest_node_id);

CREATE TABLE quest_node_children (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_node_id UUID NOT NULL REFERENCES quest_nodes(id) ON DELETE CASCADE,
  next_quest_node_id UUID NOT NULL REFERENCES quest_nodes(id) ON DELETE CASCADE,
  quest_node_challenge_id UUID REFERENCES quest_node_challenges(id) ON DELETE SET NULL
);

CREATE INDEX quest_node_children_node_idx ON quest_node_children(quest_node_id);
CREATE INDEX quest_node_children_next_idx ON quest_node_children(next_quest_node_id);

CREATE TABLE quest_acceptances_v2 (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  accepted_at TIMESTAMP NOT NULL,
  turned_in_at TIMESTAMP
);

CREATE INDEX quest_acceptances_v2_user_idx ON quest_acceptances_v2(user_id);
CREATE INDEX quest_acceptances_v2_quest_idx ON quest_acceptances_v2(quest_id);

CREATE TABLE quest_node_progress (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_acceptance_id UUID NOT NULL REFERENCES quest_acceptances_v2(id) ON DELETE CASCADE,
  quest_node_id UUID NOT NULL REFERENCES quest_nodes(id) ON DELETE CASCADE,
  completed_at TIMESTAMP
);

CREATE INDEX quest_node_progress_acceptance_idx ON quest_node_progress(quest_acceptance_id);
CREATE INDEX quest_node_progress_node_idx ON quest_node_progress(quest_node_id);
