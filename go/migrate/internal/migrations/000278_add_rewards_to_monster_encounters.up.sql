ALTER TABLE monster_encounters
  ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
  ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
  ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS reward_gold INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS item_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb;

WITH encounter_reward_stats AS (
  SELECT
    me.id AS encounter_id,
    BOOL_OR(COALESCE(m.reward_mode, 'random') = 'random') AS has_random_reward,
    MAX(
      CASE COALESCE(m.random_reward_size, 'small')
        WHEN 'large' THEN 3
        WHEN 'medium' THEN 2
        ELSE 1
      END
    ) AS max_random_reward_size_rank,
    COALESCE(
      SUM(
        CASE
          WHEN COALESCE(m.reward_mode, 'random') = 'explicit'
          THEN GREATEST(COALESCE(m.reward_experience, 0), 0)
          ELSE 0
        END
      ),
      0
    ) AS total_reward_experience,
    COALESCE(
      SUM(
        CASE
          WHEN COALESCE(m.reward_mode, 'random') = 'explicit'
          THEN GREATEST(COALESCE(m.reward_gold, 0), 0)
          ELSE 0
        END
      ),
      0
    ) AS total_reward_gold
  FROM monster_encounters me
  JOIN monster_encounter_members mem ON mem.monster_encounter_id = me.id
  JOIN monsters m ON m.id = mem.monster_id
  GROUP BY me.id
),
encounter_item_totals AS (
  SELECT
    me.id AS encounter_id,
    mir.inventory_item_id,
    SUM(mir.quantity) AS quantity
  FROM monster_encounters me
  JOIN monster_encounter_members mem ON mem.monster_encounter_id = me.id
  JOIN monsters m ON m.id = mem.monster_id
  JOIN monster_item_rewards mir ON mir.monster_id = m.id
  WHERE COALESCE(m.reward_mode, 'random') = 'explicit'
  GROUP BY me.id, mir.inventory_item_id
),
encounter_item_payloads AS (
  SELECT
    encounter_id,
    COALESCE(
      JSONB_AGG(
        JSONB_BUILD_OBJECT(
          'inventoryItemId', inventory_item_id,
          'quantity', quantity
        )
        ORDER BY inventory_item_id
      ),
      '[]'::jsonb
    ) AS item_rewards_json
  FROM encounter_item_totals
  GROUP BY encounter_id
)
UPDATE monster_encounters me
SET
  reward_mode = CASE
    WHEN stats.has_random_reward THEN 'random'
    ELSE 'explicit'
  END,
  random_reward_size = CASE stats.max_random_reward_size_rank
    WHEN 3 THEN 'large'
    WHEN 2 THEN 'medium'
    ELSE 'small'
  END,
  reward_experience = CASE
    WHEN stats.has_random_reward THEN 0
    ELSE stats.total_reward_experience
  END,
  reward_gold = CASE
    WHEN stats.has_random_reward THEN 0
    ELSE stats.total_reward_gold
  END,
  item_rewards_json = CASE
    WHEN stats.has_random_reward THEN '[]'::jsonb
    ELSE COALESCE(items.item_rewards_json, '[]'::jsonb)
  END
FROM encounter_reward_stats stats
LEFT JOIN encounter_item_payloads items ON items.encounter_id = stats.encounter_id
WHERE me.id = stats.encounter_id;

ALTER TABLE monster_encounters DROP CONSTRAINT IF EXISTS monster_encounters_reward_mode_check;
ALTER TABLE monster_encounters ADD CONSTRAINT monster_encounters_reward_mode_check CHECK (reward_mode IN ('explicit', 'random'));
ALTER TABLE monster_encounters DROP CONSTRAINT IF EXISTS monster_encounters_random_reward_size_check;
ALTER TABLE monster_encounters ADD CONSTRAINT monster_encounters_random_reward_size_check CHECK (random_reward_size IN ('small', 'medium', 'large'));
