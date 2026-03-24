ALTER TABLE quest_archetype_nodes
  ALTER COLUMN location_archetype_id DROP NOT NULL;

ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS node_type TEXT NOT NULL DEFAULT 'location',
  ADD COLUMN IF NOT EXISTS monster_ids JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS target_level INT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS encounter_reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS encounter_random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS encounter_reward_experience INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS encounter_reward_gold INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS encounter_material_rewards_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS encounter_item_rewards_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS encounter_proximity_meters INT NOT NULL DEFAULT 100;

UPDATE quest_archetype_nodes
SET
  node_type = 'location',
  monster_ids = COALESCE(monster_ids, '[]'::jsonb),
  target_level = CASE WHEN target_level < 1 THEN 1 ELSE target_level END,
  encounter_reward_mode = COALESCE(NULLIF(encounter_reward_mode, ''), 'random'),
  encounter_random_reward_size = COALESCE(NULLIF(encounter_random_reward_size, ''), 'small'),
  encounter_reward_experience = GREATEST(encounter_reward_experience, 0),
  encounter_reward_gold = GREATEST(encounter_reward_gold, 0),
  encounter_material_rewards_json = COALESCE(encounter_material_rewards_json, '[]'::jsonb),
  encounter_item_rewards_json = COALESCE(encounter_item_rewards_json, '[]'::jsonb),
  encounter_proximity_meters = GREATEST(encounter_proximity_meters, 0)
WHERE TRUE;
