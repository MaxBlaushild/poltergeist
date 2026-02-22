ALTER TABLE quest_archetype_challenges
  ADD COLUMN IF NOT EXISTS inventory_item_id INTEGER,
  ADD COLUMN IF NOT EXISTS proficiency TEXT;

CREATE TABLE quest_archetype_item_rewards (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_archetype_id UUID NOT NULL REFERENCES quest_archetypes(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX quest_archetype_item_rewards_archetype_idx ON quest_archetype_item_rewards(quest_archetype_id);
CREATE INDEX quest_archetype_item_rewards_item_idx ON quest_archetype_item_rewards(inventory_item_id);
