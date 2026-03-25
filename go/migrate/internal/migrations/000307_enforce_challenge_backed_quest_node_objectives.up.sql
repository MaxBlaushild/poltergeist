CREATE TABLE IF NOT EXISTS quest_node_challenge_legacy_archive (
  id UUID PRIMARY KEY,
  archived_at TIMESTAMP NOT NULL DEFAULT NOW(),
  quest_node_id UUID NOT NULL,
  standalone_challenge_id UUID,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  tier INTEGER NOT NULL,
  question TEXT NOT NULL,
  reward INTEGER NOT NULL,
  inventory_item_id INTEGER,
  submission_type TEXT NOT NULL DEFAULT 'photo',
  scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE,
  difficulty INTEGER NOT NULL DEFAULT 0,
  stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency TEXT,
  challenge_shuffle_status TEXT NOT NULL DEFAULT 'idle',
  challenge_shuffle_error TEXT
);

CREATE INDEX IF NOT EXISTS idx_quest_node_challenge_legacy_archive_node_id
  ON quest_node_challenge_legacy_archive(quest_node_id);

CREATE TABLE IF NOT EXISTS quest_node_child_legacy_challenge_link_archive (
  quest_node_child_id UUID PRIMARY KEY,
  quest_node_challenge_id UUID NOT NULL,
  archived_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO quest_node_child_legacy_challenge_link_archive (
  quest_node_child_id,
  quest_node_challenge_id
)
SELECT
  child.id,
  child.quest_node_challenge_id
FROM quest_node_children child
JOIN quest_node_challenges qnc ON qnc.id = child.quest_node_challenge_id
JOIN quest_nodes qn ON qn.id = qnc.quest_node_id
WHERE qn.challenge_id IS NOT NULL
ON CONFLICT (quest_node_child_id) DO NOTHING;

INSERT INTO quest_node_challenge_legacy_archive (
  id,
  quest_node_id,
  standalone_challenge_id,
  created_at,
  updated_at,
  tier,
  question,
  reward,
  inventory_item_id,
  submission_type,
  scale_with_user_level,
  difficulty,
  stat_tags,
  proficiency,
  challenge_shuffle_status,
  challenge_shuffle_error
)
SELECT
  qnc.id,
  qnc.quest_node_id,
  qn.challenge_id,
  qnc.created_at,
  qnc.updated_at,
  qnc.tier,
  qnc.question,
  qnc.reward,
  qnc.inventory_item_id,
  qnc.submission_type,
  qnc.scale_with_user_level,
  qnc.difficulty,
  COALESCE(qnc.stat_tags, '[]'::jsonb),
  qnc.proficiency,
  COALESCE(NULLIF(BTRIM(qnc.challenge_shuffle_status), ''), 'idle'),
  qnc.challenge_shuffle_error
FROM quest_node_challenges qnc
JOIN quest_nodes qn ON qn.id = qnc.quest_node_id
WHERE qn.challenge_id IS NOT NULL
ON CONFLICT (id) DO NOTHING;

WITH ranked_challenges AS (
  SELECT
    qnc.quest_node_id,
    qn.challenge_id,
    qnc.question,
    qnc.reward,
    qnc.inventory_item_id,
    qnc.submission_type,
    qnc.scale_with_user_level,
    qnc.difficulty,
    COALESCE(qnc.stat_tags, '[]'::jsonb) AS stat_tags,
    qnc.proficiency,
    ROW_NUMBER() OVER (
      PARTITION BY qnc.quest_node_id
      ORDER BY qnc.tier ASC, qnc.created_at ASC, qnc.id ASC
    ) AS row_num
  FROM quest_node_challenges qnc
  JOIN quest_nodes qn ON qn.id = qnc.quest_node_id
  WHERE qn.challenge_id IS NOT NULL
),
primary_challenges AS (
  SELECT *
  FROM ranked_challenges
  WHERE row_num = 1
)
UPDATE challenges c
SET
  question = primary_challenges.question,
  reward = primary_challenges.reward,
  inventory_item_id = primary_challenges.inventory_item_id,
  submission_type = COALESCE(NULLIF(BTRIM(primary_challenges.submission_type), ''), c.submission_type, 'photo'),
  scale_with_user_level = primary_challenges.scale_with_user_level,
  difficulty = GREATEST(primary_challenges.difficulty, 0),
  stat_tags = primary_challenges.stat_tags,
  proficiency = CASE
    WHEN primary_challenges.proficiency IS NULL OR BTRIM(primary_challenges.proficiency) = '' THEN NULL
    ELSE BTRIM(primary_challenges.proficiency)
  END,
  reward_mode = CASE
    WHEN COALESCE(primary_challenges.reward, 0) > 0 OR primary_challenges.inventory_item_id IS NOT NULL THEN 'explicit'
    ELSE c.reward_mode
  END,
  updated_at = NOW()
FROM primary_challenges
WHERE c.id = primary_challenges.challenge_id;

WITH ranked_challenges AS (
  SELECT
    qnc.quest_node_id,
    qnc.submission_type,
    ROW_NUMBER() OVER (
      PARTITION BY qnc.quest_node_id
      ORDER BY qnc.tier ASC, qnc.created_at ASC, qnc.id ASC
    ) AS row_num
  FROM quest_node_challenges qnc
  JOIN quest_nodes qn ON qn.id = qnc.quest_node_id
  WHERE qn.challenge_id IS NOT NULL
),
primary_challenges AS (
  SELECT *
  FROM ranked_challenges
  WHERE row_num = 1
)
UPDATE quest_nodes qn
SET
  submission_type = COALESCE(NULLIF(BTRIM(primary_challenges.submission_type), ''), qn.submission_type, 'photo'),
  updated_at = NOW()
FROM primary_challenges
WHERE qn.id = primary_challenges.quest_node_id;

DELETE FROM quest_node_challenges qnc
USING quest_nodes qn
WHERE qnc.quest_node_id = qn.id
  AND qn.challenge_id IS NOT NULL;

CREATE OR REPLACE FUNCTION prevent_quest_node_challenges_on_challenge_nodes()
RETURNS TRIGGER AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM quest_nodes
    WHERE id = NEW.quest_node_id
      AND challenge_id IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'challenge-backed quest nodes cannot have quest_node_challenges';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS quest_node_challenges_prevent_challenge_targets
  ON quest_node_challenges;

CREATE TRIGGER quest_node_challenges_prevent_challenge_targets
BEFORE INSERT OR UPDATE OF quest_node_id
ON quest_node_challenges
FOR EACH ROW
EXECUTE FUNCTION prevent_quest_node_challenges_on_challenge_nodes();

CREATE OR REPLACE FUNCTION prevent_challenge_targets_on_nodes_with_prompts()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.challenge_id IS NOT NULL AND EXISTS (
    SELECT 1
    FROM quest_node_challenges
    WHERE quest_node_id = NEW.id
  ) THEN
    RAISE EXCEPTION 'quest nodes with challenge_id cannot have quest_node_challenges';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS quest_nodes_prevent_prompt_overlap
  ON quest_nodes;

CREATE TRIGGER quest_nodes_prevent_prompt_overlap
BEFORE INSERT OR UPDATE OF challenge_id
ON quest_nodes
FOR EACH ROW
EXECUTE FUNCTION prevent_challenge_targets_on_nodes_with_prompts();
