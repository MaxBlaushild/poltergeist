DROP TRIGGER IF EXISTS quest_node_challenges_prevent_challenge_targets
  ON quest_node_challenges;

DROP FUNCTION IF EXISTS prevent_quest_node_challenges_on_challenge_nodes();

DROP TRIGGER IF EXISTS quest_nodes_prevent_prompt_overlap
  ON quest_nodes;

DROP FUNCTION IF EXISTS prevent_challenge_targets_on_nodes_with_prompts();

INSERT INTO quest_node_challenges (
  id,
  created_at,
  updated_at,
  quest_node_id,
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
  archive.id,
  archive.created_at,
  archive.updated_at,
  archive.quest_node_id,
  archive.tier,
  archive.question,
  archive.reward,
  archive.inventory_item_id,
  archive.submission_type,
  archive.scale_with_user_level,
  archive.difficulty,
  archive.stat_tags,
  archive.proficiency,
  archive.challenge_shuffle_status,
  archive.challenge_shuffle_error
FROM quest_node_challenge_legacy_archive archive
WHERE EXISTS (
    SELECT 1
    FROM quest_nodes qn
    WHERE qn.id = archive.quest_node_id
  )
  AND NOT EXISTS (
    SELECT 1
    FROM quest_node_challenges qnc
    WHERE qnc.id = archive.id
  );

UPDATE quest_node_children child
SET quest_node_challenge_id = archive.quest_node_challenge_id
FROM quest_node_child_legacy_challenge_link_archive archive
WHERE child.id = archive.quest_node_child_id
  AND child.quest_node_challenge_id IS NULL
  AND EXISTS (
    SELECT 1
    FROM quest_node_challenges qnc
    WHERE qnc.id = archive.quest_node_challenge_id
  );

DROP TABLE IF EXISTS quest_node_child_legacy_challenge_link_archive;

DROP INDEX IF EXISTS idx_quest_node_challenge_legacy_archive_node_id;

DROP TABLE IF EXISTS quest_node_challenge_legacy_archive;
