ALTER TABLE quest_node_challenges
  ADD COLUMN challenge_shuffle_status TEXT NOT NULL DEFAULT 'idle',
  ADD COLUMN challenge_shuffle_error TEXT;
