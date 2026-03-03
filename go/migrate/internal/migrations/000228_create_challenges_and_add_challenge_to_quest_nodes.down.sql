DROP INDEX IF EXISTS quest_nodes_challenge_idx;

ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS challenge_id;

DROP INDEX IF EXISTS idx_challenges_geometry;
DROP INDEX IF EXISTS idx_challenges_zone_id;
DROP TABLE IF EXISTS challenges;
