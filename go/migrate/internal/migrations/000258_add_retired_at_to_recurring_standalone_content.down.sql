DROP INDEX IF EXISTS monster_encounters_retired_at_idx;

ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS retired_at;

DROP INDEX IF EXISTS challenges_retired_at_idx;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS retired_at;

DROP INDEX IF EXISTS scenarios_retired_at_idx;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS retired_at;
