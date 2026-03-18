ALTER TABLE treasure_chests DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE monster_encounters DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE monsters DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE challenges DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE scenario_options DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE scenarios DROP COLUMN IF EXISTS material_rewards_json;
ALTER TABLE quests DROP COLUMN IF EXISTS material_rewards_json;
