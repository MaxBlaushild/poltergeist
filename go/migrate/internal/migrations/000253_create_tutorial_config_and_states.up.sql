ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS ephemeral BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_scenarios_owner_user_id ON scenarios(owner_user_id);

CREATE TABLE IF NOT EXISTS tutorial_configs (
  id INTEGER PRIMARY KEY,
  character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
  dialogue_json JSONB NOT NULL DEFAULT '[]',
  scenario_prompt TEXT NOT NULL DEFAULT 'You hear a commotion outside of your door.',
  scenario_image_url TEXT NOT NULL DEFAULT '',
  options_json JSONB NOT NULL DEFAULT '[]',
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  item_rewards_json JSONB NOT NULL DEFAULT '[]',
  spell_rewards_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tutorial_configs (
  id,
  dialogue_json,
  scenario_prompt,
  scenario_image_url,
  options_json,
  reward_experience,
  reward_gold,
  item_rewards_json,
  spell_rewards_json
) VALUES (
  1,
  '["Welcome, traveler, to these unclaimed streets!","We have plenty of problems here. But opportunity as well.","Explore, grow, and conquer!"]',
  'You hear a commotion outside of your door.',
  '',
  '[{"optionText":"I reach for my sword and check it out.","statTag":"strength"},{"optionText":"I reach for my shield and check it out.","statTag":"constitution"},{"optionText":"I reach for my spellbook and check it out.","statTag":"intelligence"}]',
  0,
  0,
  '[]',
  '[]'
) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_tutorial_states (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  tutorial_scenario_id UUID REFERENCES scenarios(id) ON DELETE SET NULL,
  activated_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_tutorial_states_scenario_id
  ON user_tutorial_states(tutorial_scenario_id)
  WHERE tutorial_scenario_id IS NOT NULL;
