CREATE TABLE IF NOT EXISTS quest_node_challenges (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_node_id UUID NOT NULL REFERENCES quest_nodes(id) ON DELETE CASCADE,
  tier INTEGER NOT NULL,
  question TEXT NOT NULL,
  reward INTEGER NOT NULL DEFAULT 0,
  inventory_item_id INTEGER,
  submission_type TEXT NOT NULL DEFAULT 'photo',
  scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE,
  difficulty INTEGER NOT NULL DEFAULT 0,
  stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency TEXT,
  challenge_shuffle_status TEXT NOT NULL DEFAULT 'idle',
  challenge_shuffle_error TEXT
);

CREATE INDEX IF NOT EXISTS quest_node_challenges_node_idx
  ON quest_node_challenges(quest_node_id);

ALTER TABLE quest_node_children
  ADD COLUMN IF NOT EXISTS quest_node_challenge_id UUID REFERENCES quest_node_challenges(id) ON DELETE SET NULL;
