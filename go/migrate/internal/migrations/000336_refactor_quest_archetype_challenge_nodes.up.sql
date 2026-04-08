ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS challenge_template_id UUID,
  ADD COLUMN IF NOT EXISTS location_selection_mode TEXT NOT NULL DEFAULT 'random';

UPDATE quest_archetype_nodes AS nodes
SET challenge_template_id = link.challenge_template_id
FROM (
  SELECT DISTINCT ON (joins.quest_archetype_node_id)
    joins.quest_archetype_node_id,
    challenges.challenge_template_id
  FROM quest_archetype_node_challenges AS joins
  INNER JOIN quest_archetype_challenges AS challenges
    ON challenges.id = joins.quest_archetype_challenge_id
  WHERE challenges.challenge_template_id IS NOT NULL
  ORDER BY joins.quest_archetype_node_id, challenges.created_at ASC, challenges.id ASC
) AS link
WHERE nodes.id = link.quest_archetype_node_id
  AND nodes.challenge_template_id IS NULL;

UPDATE quest_archetype_nodes
SET
  node_type = CASE
    WHEN node_type = 'location' THEN 'challenge'
    ELSE node_type
  END,
  location_selection_mode = COALESCE(NULLIF(location_selection_mode, ''), 'random')
WHERE TRUE;
