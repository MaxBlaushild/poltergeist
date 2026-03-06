CREATE TABLE user_healing_fountain_discoveries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  healing_fountain_id UUID NOT NULL REFERENCES healing_fountains(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_healing_fountain_discoveries_user_fountain
  ON user_healing_fountain_discoveries(user_id, healing_fountain_id);

CREATE INDEX idx_user_healing_fountain_discoveries_user_id
  ON user_healing_fountain_discoveries(user_id);

CREATE INDEX idx_user_healing_fountain_discoveries_fountain_id
  ON user_healing_fountain_discoveries(healing_fountain_id);
