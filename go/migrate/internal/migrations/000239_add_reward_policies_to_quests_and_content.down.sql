ALTER TABLE quests DROP CONSTRAINT IF EXISTS quests_reward_mode_check;
ALTER TABLE quests DROP CONSTRAINT IF EXISTS quests_random_reward_size_check;
ALTER TABLE scenarios DROP CONSTRAINT IF EXISTS scenarios_reward_mode_check;
ALTER TABLE scenarios DROP CONSTRAINT IF EXISTS scenarios_random_reward_size_check;
ALTER TABLE challenges DROP CONSTRAINT IF EXISTS challenges_reward_mode_check;
ALTER TABLE challenges DROP CONSTRAINT IF EXISTS challenges_random_reward_size_check;
ALTER TABLE monsters DROP CONSTRAINT IF EXISTS monsters_reward_mode_check;
ALTER TABLE monsters DROP CONSTRAINT IF EXISTS monsters_random_reward_size_check;

ALTER TABLE quests
  DROP COLUMN IF EXISTS reward_mode,
  DROP COLUMN IF EXISTS random_reward_size,
  DROP COLUMN IF EXISTS reward_experience;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS reward_mode,
  DROP COLUMN IF EXISTS random_reward_size;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS reward_mode,
  DROP COLUMN IF EXISTS random_reward_size,
  DROP COLUMN IF EXISTS reward_experience;

ALTER TABLE monsters
  DROP COLUMN IF EXISTS reward_mode,
  DROP COLUMN IF EXISTS random_reward_size;
