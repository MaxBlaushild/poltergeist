DROP INDEX IF EXISTS idx_challenges_polygon;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS polygon;
