ALTER TABLE quests
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE scenarios
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE scenario_options
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE challenges
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE monsters
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE monster_encounters
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE treasure_chests
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;
