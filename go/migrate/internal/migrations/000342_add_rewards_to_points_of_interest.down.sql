DROP INDEX IF EXISTS point_of_interest_spell_rewards_point_of_interest_id_idx;
DROP TABLE IF EXISTS point_of_interest_spell_rewards;

DROP INDEX IF EXISTS point_of_interest_item_rewards_point_of_interest_id_idx;
DROP TABLE IF EXISTS point_of_interest_item_rewards;

ALTER TABLE points_of_interest DROP CONSTRAINT IF EXISTS points_of_interest_random_reward_size_check;
ALTER TABLE points_of_interest DROP CONSTRAINT IF EXISTS points_of_interest_reward_mode_check;

ALTER TABLE points_of_interest
  DROP COLUMN IF EXISTS material_rewards_json,
  DROP COLUMN IF EXISTS reward_gold,
  DROP COLUMN IF EXISTS reward_experience,
  DROP COLUMN IF EXISTS random_reward_size,
  DROP COLUMN IF EXISTS reward_mode;
