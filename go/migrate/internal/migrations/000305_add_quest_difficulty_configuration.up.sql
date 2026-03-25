ALTER TABLE quest_archetypes
  ADD COLUMN IF NOT EXISTS difficulty_mode TEXT NOT NULL DEFAULT 'fixed',
  ADD COLUMN IF NOT EXISTS difficulty INTEGER NOT NULL DEFAULT 1;

ALTER TABLE quests
  ADD COLUMN IF NOT EXISTS difficulty_mode TEXT NOT NULL DEFAULT 'fixed',
  ADD COLUMN IF NOT EXISTS difficulty INTEGER NOT NULL DEFAULT 1;

ALTER TABLE quest_node_challenges
  ADD COLUMN IF NOT EXISTS scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE quest_archetypes qa
SET difficulty = GREATEST(1, COALESCE(root.difficulty, 1))
FROM quest_archetype_nodes root
WHERE qa.root_id = root.id;

UPDATE quests q
SET difficulty = GREATEST(1, COALESCE(agg.avg_difficulty, 1))
FROM (
  SELECT qn.quest_id, ROUND(AVG(qnc.difficulty))::INTEGER AS avg_difficulty
  FROM quest_nodes qn
  JOIN quest_node_challenges qnc ON qnc.quest_node_id = qn.id
  GROUP BY qn.quest_id
) agg
WHERE q.id = agg.quest_id;

UPDATE quest_archetypes
SET difficulty = GREATEST(1, difficulty);

UPDATE quests
SET difficulty = GREATEST(1, difficulty);
