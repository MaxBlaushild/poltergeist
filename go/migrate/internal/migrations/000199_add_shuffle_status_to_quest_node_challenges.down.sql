ALTER TABLE quest_node_challenges
  DROP COLUMN IF EXISTS challenge_shuffle_error,
  DROP COLUMN IF EXISTS challenge_shuffle_status;
