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
WHERE qn.scenario_id IS NOT NULL
   OR qn.monster_encounter_id IS NOT NULL
   OR qn.monster_id IS NOT NULL
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
  NULL,
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
WHERE qn.scenario_id IS NOT NULL
   OR qn.monster_encounter_id IS NOT NULL
   OR qn.monster_id IS NOT NULL
ON CONFLICT (id) DO NOTHING;

DELETE FROM quest_node_challenges qnc
USING quest_nodes qn
WHERE qnc.quest_node_id = qn.id
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
      AND (
        challenge_id IS NOT NULL OR
        scenario_id IS NOT NULL OR
        monster_encounter_id IS NOT NULL OR
        monster_id IS NOT NULL
      )
  ) THEN
    RAISE EXCEPTION 'target-backed quest nodes cannot have quest_node_challenges';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION prevent_challenge_targets_on_nodes_with_prompts()
RETURNS TRIGGER AS $$
BEGIN
  IF (
    NEW.challenge_id IS NOT NULL OR
    NEW.scenario_id IS NOT NULL OR
    NEW.monster_encounter_id IS NOT NULL OR
    NEW.monster_id IS NOT NULL
  ) AND EXISTS (
    SELECT 1
    FROM quest_node_challenges
    WHERE quest_node_id = NEW.id
  ) THEN
    RAISE EXCEPTION 'target-backed quest nodes cannot have quest_node_challenges';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
