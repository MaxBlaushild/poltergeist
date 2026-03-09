CREATE TABLE challenge_item_choice_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX idx_challenge_item_choice_rewards_challenge_id ON challenge_item_choice_rewards(challenge_id);

CREATE TABLE scenario_item_choice_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX idx_scenario_item_choice_rewards_scenario_id ON scenario_item_choice_rewards(scenario_id);

CREATE TABLE scenario_option_item_choice_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_option_id UUID NOT NULL REFERENCES scenario_options(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX idx_scenario_option_item_choice_rewards_option_id ON scenario_option_item_choice_rewards(scenario_option_id);

CREATE TABLE user_challenge_item_choice_pendings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  UNIQUE(user_id, challenge_id)
);

CREATE INDEX idx_user_challenge_item_choice_pendings_user_id ON user_challenge_item_choice_pendings(user_id);

CREATE TABLE user_scenario_item_choice_pendings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  scenario_option_id UUID REFERENCES scenario_options(id) ON DELETE SET NULL,
  UNIQUE(user_id, scenario_id)
);

CREATE INDEX idx_user_scenario_item_choice_pendings_user_id ON user_scenario_item_choice_pendings(user_id);
