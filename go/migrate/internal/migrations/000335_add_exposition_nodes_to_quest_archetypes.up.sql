ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS exposition_title TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS exposition_description TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS exposition_dialogue JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS exposition_reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS exposition_random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS exposition_reward_experience INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS exposition_reward_gold INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS exposition_material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS exposition_item_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS exposition_spell_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE quest_archetype_nodes
SET
  exposition_title = COALESCE(exposition_title, ''),
  exposition_description = COALESCE(exposition_description, ''),
  exposition_dialogue = COALESCE(exposition_dialogue, '[]'::jsonb),
  exposition_reward_mode = COALESCE(NULLIF(exposition_reward_mode, ''), 'random'),
  exposition_random_reward_size = COALESCE(NULLIF(exposition_random_reward_size, ''), 'small'),
  exposition_reward_experience = GREATEST(exposition_reward_experience, 0),
  exposition_reward_gold = GREATEST(exposition_reward_gold, 0),
  exposition_material_rewards_json = COALESCE(exposition_material_rewards_json, '[]'::jsonb),
  exposition_item_rewards_json = COALESCE(exposition_item_rewards_json, '[]'::jsonb),
  exposition_spell_rewards_json = COALESCE(exposition_spell_rewards_json, '[]'::jsonb)
WHERE TRUE;
