ALTER TABLE quests
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0;

ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small';

ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0;

ALTER TABLE monsters
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small';

UPDATE quests
SET reward_mode = 'explicit'
WHERE
  gold > 0
  OR reward_experience > 0
  OR EXISTS (
    SELECT 1
    FROM quest_item_rewards qir
    WHERE qir.quest_id = quests.id
  )
  OR EXISTS (
    SELECT 1
    FROM quest_spell_rewards qsr
    WHERE qsr.quest_id = quests.id
  );

UPDATE scenarios
SET reward_mode = 'explicit'
WHERE
  reward_experience > 0
  OR reward_gold > 0
  OR EXISTS (
    SELECT 1
    FROM scenario_item_rewards sir
    WHERE sir.scenario_id = scenarios.id
  )
  OR EXISTS (
    SELECT 1
    FROM scenario_spell_rewards ssr
    WHERE ssr.scenario_id = scenarios.id
  )
  OR EXISTS (
    SELECT 1
    FROM scenario_options so
    WHERE so.scenario_id = scenarios.id
      AND (
        so.reward_experience > 0
        OR so.reward_gold > 0
        OR EXISTS (
          SELECT 1
          FROM scenario_option_item_rewards soir
          WHERE soir.scenario_option_id = so.id
        )
        OR EXISTS (
          SELECT 1
          FROM scenario_option_spell_rewards sosr
          WHERE sosr.scenario_option_id = so.id
        )
      )
  );

UPDATE challenges
SET reward_mode = 'explicit'
WHERE reward > 0 OR reward_experience > 0 OR inventory_item_id IS NOT NULL;

UPDATE monsters
SET reward_mode = 'explicit'
WHERE
  reward_experience > 0
  OR reward_gold > 0
  OR EXISTS (
    SELECT 1
    FROM monster_item_rewards mir
    WHERE mir.monster_id = monsters.id
  );

ALTER TABLE quests DROP CONSTRAINT IF EXISTS quests_reward_mode_check;
ALTER TABLE quests ADD CONSTRAINT quests_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE quests DROP CONSTRAINT IF EXISTS quests_random_reward_size_check;
ALTER TABLE quests ADD CONSTRAINT quests_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));

ALTER TABLE scenarios DROP CONSTRAINT IF EXISTS scenarios_reward_mode_check;
ALTER TABLE scenarios ADD CONSTRAINT scenarios_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE scenarios DROP CONSTRAINT IF EXISTS scenarios_random_reward_size_check;
ALTER TABLE scenarios ADD CONSTRAINT scenarios_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));

ALTER TABLE challenges DROP CONSTRAINT IF EXISTS challenges_reward_mode_check;
ALTER TABLE challenges ADD CONSTRAINT challenges_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE challenges DROP CONSTRAINT IF EXISTS challenges_random_reward_size_check;
ALTER TABLE challenges ADD CONSTRAINT challenges_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));

ALTER TABLE monsters DROP CONSTRAINT IF EXISTS monsters_reward_mode_check;
ALTER TABLE monsters ADD CONSTRAINT monsters_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE monsters DROP CONSTRAINT IF EXISTS monsters_random_reward_size_check;
ALTER TABLE monsters ADD CONSTRAINT monsters_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));
