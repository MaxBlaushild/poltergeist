ALTER TABLE points_of_interest
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS reward_gold INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE points_of_interest
SET reward_mode = COALESCE(NULLIF(reward_mode, ''), 'random'),
    random_reward_size = COALESCE(NULLIF(random_reward_size, ''), 'small'),
    reward_experience = GREATEST(COALESCE(reward_experience, 0), 0),
    reward_gold = GREATEST(COALESCE(reward_gold, 0), 0),
    material_rewards_json = COALESCE(material_rewards_json, '[]'::jsonb);

ALTER TABLE points_of_interest DROP CONSTRAINT IF EXISTS points_of_interest_reward_mode_check;
ALTER TABLE points_of_interest
  ADD CONSTRAINT points_of_interest_reward_mode_check
  CHECK (reward_mode IN ('explicit', 'random'));

ALTER TABLE points_of_interest DROP CONSTRAINT IF EXISTS points_of_interest_random_reward_size_check;
ALTER TABLE points_of_interest
  ADD CONSTRAINT points_of_interest_random_reward_size_check
  CHECK (random_reward_size IN ('small', 'medium', 'large'));

CREATE TABLE IF NOT EXISTS point_of_interest_item_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS point_of_interest_item_rewards_point_of_interest_id_idx
  ON point_of_interest_item_rewards(point_of_interest_id);

CREATE TABLE IF NOT EXISTS point_of_interest_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  point_of_interest_id UUID NOT NULL REFERENCES points_of_interest(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id)
);

CREATE INDEX IF NOT EXISTS point_of_interest_spell_rewards_point_of_interest_id_idx
  ON point_of_interest_spell_rewards(point_of_interest_id);
