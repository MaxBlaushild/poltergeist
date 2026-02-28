CREATE TABLE spells (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  icon_url TEXT NOT NULL DEFAULT '',
  effect_text TEXT NOT NULL DEFAULT '',
  school_of_magic TEXT NOT NULL DEFAULT '',
  mana_cost INTEGER NOT NULL DEFAULT 0,
  effects JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX idx_spells_name ON spells(name);

CREATE TABLE user_spells (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  acquired_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, spell_id)
);

CREATE INDEX idx_user_spells_user_id ON user_spells(user_id);
CREATE INDEX idx_user_spells_spell_id ON user_spells(spell_id);

CREATE TABLE quest_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  UNIQUE(quest_id, spell_id)
);

CREATE INDEX idx_quest_spell_rewards_quest_id ON quest_spell_rewards(quest_id);
CREATE INDEX idx_quest_spell_rewards_spell_id ON quest_spell_rewards(spell_id);

CREATE TABLE scenario_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  UNIQUE(scenario_id, spell_id)
);

CREATE INDEX idx_scenario_spell_rewards_scenario_id ON scenario_spell_rewards(scenario_id);
CREATE INDEX idx_scenario_spell_rewards_spell_id ON scenario_spell_rewards(spell_id);

CREATE TABLE scenario_option_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  scenario_option_id UUID NOT NULL REFERENCES scenario_options(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  UNIQUE(scenario_option_id, spell_id)
);

CREATE INDEX idx_scenario_option_spell_rewards_option_id ON scenario_option_spell_rewards(scenario_option_id);
CREATE INDEX idx_scenario_option_spell_rewards_spell_id ON scenario_option_spell_rewards(spell_id);
