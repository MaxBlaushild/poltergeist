ALTER TABLE quest_acceptances_v2
  ADD COLUMN IF NOT EXISTS current_quest_node_id UUID;

CREATE INDEX IF NOT EXISTS quest_acceptances_v2_current_node_idx
  ON quest_acceptances_v2(current_quest_node_id);

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS failure_policy TEXT NOT NULL DEFAULT 'retry';

UPDATE quest_nodes
SET failure_policy = 'retry'
WHERE failure_policy IS NULL OR BTRIM(failure_policy) = '';

ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS failure_policy TEXT NOT NULL DEFAULT 'retry';

UPDATE quest_archetype_nodes
SET failure_policy = 'retry'
WHERE failure_policy IS NULL OR BTRIM(failure_policy) = '';

ALTER TABLE quest_archetype_challenges
  ADD COLUMN IF NOT EXISTS failure_unlocked_node_id UUID;

CREATE INDEX IF NOT EXISTS quest_archetype_challenges_failure_unlocked_idx
  ON quest_archetype_challenges(failure_unlocked_node_id);

ALTER TABLE quest_node_children
  ADD COLUMN IF NOT EXISTS outcome TEXT NOT NULL DEFAULT 'success';

UPDATE quest_node_children
SET outcome = 'success'
WHERE outcome IS NULL OR BTRIM(outcome) = '';

ALTER TABLE quest_node_progress
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS attempt_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_failed_at TIMESTAMP,
  ADD COLUMN IF NOT EXISTS last_failure_reason TEXT NOT NULL DEFAULT '';

UPDATE quest_node_progress
SET status = CASE
  WHEN completed_at IS NOT NULL THEN 'completed'
  ELSE 'active'
END,
attempt_count = CASE
  WHEN completed_at IS NOT NULL AND attempt_count = 0 THEN 1
  ELSE attempt_count
END,
last_failure_reason = COALESCE(last_failure_reason, '')
WHERE status IS NULL OR BTRIM(status) = '' OR last_failure_reason IS NULL;
