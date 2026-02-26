CREATE TABLE scenarios (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326) NOT NULL,
  prompt TEXT NOT NULL,
  image_url TEXT NOT NULL,
  thumbnail_url TEXT NOT NULL,
  difficulty INTEGER NOT NULL DEFAULT 24,
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  open_ended BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_scenarios_zone_id ON scenarios(zone_id);
CREATE INDEX idx_scenarios_geometry ON scenarios USING GIST(geometry);

CREATE TABLE scenario_options (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  option_text TEXT NOT NULL,
  stat_tag TEXT NOT NULL,
  proficiencies JSONB NOT NULL DEFAULT '[]'::jsonb,
  difficulty INTEGER,
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_scenario_options_scenario_id ON scenario_options(scenario_id);

CREATE TABLE scenario_option_item_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_option_id UUID NOT NULL REFERENCES scenario_options(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_scenario_option_item_rewards_option_id ON scenario_option_item_rewards(scenario_option_id);

CREATE TABLE scenario_item_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_scenario_item_rewards_scenario_id ON scenario_item_rewards(scenario_id);

CREATE TABLE user_scenario_attempts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  scenario_option_id UUID REFERENCES scenario_options(id) ON DELETE SET NULL,
  freeform_response TEXT,
  roll INTEGER NOT NULL,
  stat_tag TEXT NOT NULL,
  stat_value INTEGER NOT NULL DEFAULT 0,
  proficiencies_used JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency_bonus INTEGER NOT NULL DEFAULT 0,
  creativity_bonus INTEGER NOT NULL DEFAULT 0,
  threshold INTEGER NOT NULL,
  total_score INTEGER NOT NULL,
  successful BOOLEAN NOT NULL DEFAULT FALSE,
  reasoning TEXT,
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  UNIQUE(user_id, scenario_id)
);

CREATE INDEX idx_user_scenario_attempts_user_id ON user_scenario_attempts(user_id);
CREATE INDEX idx_user_scenario_attempts_scenario_id ON user_scenario_attempts(scenario_id);
