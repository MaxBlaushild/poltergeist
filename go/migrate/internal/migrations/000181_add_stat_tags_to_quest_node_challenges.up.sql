ALTER TABLE quest_node_challenges
  ADD COLUMN stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
