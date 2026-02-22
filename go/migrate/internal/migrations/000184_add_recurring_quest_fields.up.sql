ALTER TABLE quests
  ADD COLUMN recurring_quest_id UUID,
  ADD COLUMN recurrence_frequency TEXT,
  ADD COLUMN next_recurrence_at TIMESTAMP;

CREATE INDEX quests_recurring_quest_id_idx ON quests(recurring_quest_id);
CREATE INDEX quests_next_recurrence_at_idx ON quests(next_recurrence_at);
