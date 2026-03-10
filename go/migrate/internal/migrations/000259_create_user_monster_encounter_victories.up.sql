CREATE TABLE user_monster_encounter_victories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  monster_encounter_id UUID NOT NULL REFERENCES monster_encounters(id) ON DELETE CASCADE,
  UNIQUE(user_id, monster_encounter_id)
);

CREATE INDEX idx_user_monster_encounter_victories_user_id
  ON user_monster_encounter_victories(user_id);

CREATE INDEX idx_user_monster_encounter_victories_encounter_id
  ON user_monster_encounter_victories(monster_encounter_id);
