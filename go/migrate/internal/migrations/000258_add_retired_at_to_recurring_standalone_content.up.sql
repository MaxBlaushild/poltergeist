ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS retired_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS scenarios_retired_at_idx
  ON scenarios(retired_at);

ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS retired_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS challenges_retired_at_idx
  ON challenges(retired_at);

ALTER TABLE monster_encounters
  ADD COLUMN IF NOT EXISTS retired_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS monster_encounters_retired_at_idx
  ON monster_encounters(retired_at);
