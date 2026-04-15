DROP INDEX IF EXISTS quest_acceptances_v2_current_node_idx;
ALTER TABLE quest_acceptances_v2
  DROP COLUMN IF EXISTS current_quest_node_id;

DROP INDEX IF EXISTS quest_archetype_challenges_failure_unlocked_idx;
ALTER TABLE quest_archetype_challenges
  DROP COLUMN IF EXISTS failure_unlocked_node_id;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS failure_policy;

ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS failure_policy;

ALTER TABLE quest_node_children
  DROP COLUMN IF EXISTS outcome;

ALTER TABLE quest_node_progress
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS attempt_count,
  DROP COLUMN IF EXISTS last_failed_at,
  DROP COLUMN IF EXISTS last_failure_reason;
