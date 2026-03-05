ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_scenarios_point_of_interest_id
  ON scenarios(point_of_interest_id);

ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_challenges_point_of_interest_id
  ON challenges(point_of_interest_id);
