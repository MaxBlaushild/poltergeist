CREATE TABLE expositions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326) NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  dialogue JSONB NOT NULL DEFAULT '[]'::jsonb,
  required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb,
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  reward_mode TEXT NOT NULL DEFAULT 'random',
  random_reward_size TEXT NOT NULL DEFAULT 'small',
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  CONSTRAINT expositions_reward_mode_check CHECK (reward_mode IN ('explicit', 'random')),
  CONSTRAINT expositions_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'))
);

CREATE INDEX idx_expositions_zone_id ON expositions(zone_id);
CREATE INDEX idx_expositions_geometry ON expositions USING GIST(geometry);
CREATE INDEX idx_expositions_point_of_interest_id ON expositions(point_of_interest_id);

CREATE TABLE exposition_item_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  exposition_id UUID NOT NULL REFERENCES expositions(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_exposition_item_rewards_exposition_id ON exposition_item_rewards(exposition_id);

CREATE TABLE exposition_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  exposition_id UUID NOT NULL REFERENCES expositions(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE
);

CREATE INDEX idx_exposition_spell_rewards_exposition_id ON exposition_spell_rewards(exposition_id);
CREATE INDEX idx_exposition_spell_rewards_spell_id ON exposition_spell_rewards(spell_id);

CREATE TABLE user_exposition_completions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exposition_id UUID NOT NULL REFERENCES expositions(id) ON DELETE CASCADE,
  UNIQUE(user_id, exposition_id)
);

CREATE INDEX idx_user_exposition_completions_user_id ON user_exposition_completions(user_id);
CREATE INDEX idx_user_exposition_completions_exposition_id ON user_exposition_completions(exposition_id);

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS exposition_id UUID REFERENCES expositions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS quest_nodes_exposition_idx ON quest_nodes(exposition_id);
