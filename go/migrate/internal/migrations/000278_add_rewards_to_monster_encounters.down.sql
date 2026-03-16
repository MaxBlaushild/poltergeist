ALTER TABLE monster_encounters DROP CONSTRAINT IF EXISTS monster_encounters_random_reward_size_check;
ALTER TABLE monster_encounters DROP CONSTRAINT IF EXISTS monster_encounters_reward_mode_check;

ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS item_rewards_json,
  DROP COLUMN IF EXISTS reward_gold,
  DROP COLUMN IF EXISTS reward_experience,
  DROP COLUMN IF EXISTS random_reward_size,
  DROP COLUMN IF EXISTS reward_mode;
