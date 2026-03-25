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
JOIN quest_nodes qn ON qn.id = archive.quest_node_id
WHERE archive.standalone_challenge_id IS NULL
  AND (
    qn.scenario_id IS NOT NULL OR
    qn.monster_encounter_id IS NOT NULL OR
    qn.monster_id IS NOT NULL
  )
  AND NOT EXISTS (
    SELECT 1
    FROM quest_node_challenges qnc
    WHERE qnc.id = archive.id
  );

UPDATE quest_node_children child
SET quest_node_challenge_id = archive.quest_node_challenge_id
FROM quest_node_child_legacy_challenge_link_archive archive
JOIN quest_node_challenges qnc ON qnc.id = archive.quest_node_challenge_id
JOIN quest_nodes qn ON qn.id = qnc.quest_node_id
WHERE child.id = archive.quest_node_child_id
  AND child.quest_node_challenge_id IS NULL
  AND (
    qn.scenario_id IS NOT NULL OR
    qn.monster_encounter_id IS NOT NULL OR
    qn.monster_id IS NOT NULL
  );

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
