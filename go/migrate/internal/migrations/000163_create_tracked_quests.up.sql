CREATE TABLE tracked_quests (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP,
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX tracked_quests_user_idx ON tracked_quests(user_id);
CREATE INDEX tracked_quests_quest_idx ON tracked_quests(quest_id);

INSERT INTO tracked_quests (id, created_at, updated_at, deleted_at, quest_id, user_id)
SELECT id, created_at, updated_at, deleted_at, point_of_interest_group_id, user_id
FROM tracked_point_of_interest_groups
WHERE deleted_at IS NULL;
