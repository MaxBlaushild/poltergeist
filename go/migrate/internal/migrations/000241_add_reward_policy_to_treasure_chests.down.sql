ALTER TABLE treasure_chests DROP CONSTRAINT IF EXISTS treasure_chests_reward_mode_check;
ALTER TABLE treasure_chests DROP CONSTRAINT IF EXISTS treasure_chests_random_reward_size_check;

ALTER TABLE treasure_chests
  DROP COLUMN IF EXISTS reward_mode,
  DROP COLUMN IF EXISTS random_reward_size,
  DROP COLUMN IF EXISTS reward_experience;
