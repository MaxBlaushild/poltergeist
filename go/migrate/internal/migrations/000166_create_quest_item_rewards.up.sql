CREATE TABLE quest_item_rewards (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX quest_item_rewards_quest_idx ON quest_item_rewards(quest_id);
CREATE INDEX quest_item_rewards_item_idx ON quest_item_rewards(inventory_item_id);
