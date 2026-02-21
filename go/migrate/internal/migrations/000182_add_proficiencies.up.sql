ALTER TABLE quest_node_challenges
ADD COLUMN proficiency TEXT;

CREATE TABLE user_proficiencies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  proficiency TEXT NOT NULL,
  level INT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX user_proficiencies_user_id_proficiency_idx
  ON user_proficiencies(user_id, proficiency);

CREATE INDEX user_proficiencies_user_id_idx
  ON user_proficiencies(user_id);
