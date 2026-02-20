ALTER TABLE quests
  ADD COLUMN stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
