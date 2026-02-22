DROP INDEX IF EXISTS quests_next_recurrence_at_idx;
DROP INDEX IF EXISTS quests_recurring_quest_id_idx;

ALTER TABLE quests
  DROP COLUMN IF EXISTS next_recurrence_at,
  DROP COLUMN IF EXISTS recurrence_frequency,
  DROP COLUMN IF EXISTS recurring_quest_id;
