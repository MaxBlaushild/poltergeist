ALTER TABLE treasure_chests
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0;

UPDATE treasure_chests
SET reward_mode = 'explicit'
WHERE
  COALESCE(gold, 0) > 0
  OR reward_experience > 0
  OR EXISTS (
    SELECT 1
    FROM treasure_chest_items tci
    WHERE tci.treasure_chest_id = treasure_chests.id
  );

ALTER TABLE treasure_chests DROP CONSTRAINT IF EXISTS treasure_chests_reward_mode_check;
ALTER TABLE treasure_chests ADD CONSTRAINT treasure_chests_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE treasure_chests DROP CONSTRAINT IF EXISTS treasure_chests_random_reward_size_check;
ALTER TABLE treasure_chests ADD CONSTRAINT treasure_chests_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));
