DROP INDEX IF EXISTS idx_challenges_point_of_interest_id;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS point_of_interest_id;

DROP INDEX IF EXISTS idx_scenarios_point_of_interest_id;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS point_of_interest_id;
