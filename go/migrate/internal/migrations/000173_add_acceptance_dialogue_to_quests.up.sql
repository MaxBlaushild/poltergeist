ALTER TABLE quests
  ADD COLUMN acceptance_dialogue JSONB NOT NULL DEFAULT '[]'::jsonb;
