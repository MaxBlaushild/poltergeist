DROP INDEX IF EXISTS monster_encounters_next_recurrence_at_idx;
DROP INDEX IF EXISTS monster_encounters_recurring_monster_encounter_id_idx;

ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS next_recurrence_at,
  DROP COLUMN IF EXISTS recurrence_frequency,
  DROP COLUMN IF EXISTS recurring_monster_encounter_id;

DROP INDEX IF EXISTS challenges_next_recurrence_at_idx;
DROP INDEX IF EXISTS challenges_recurring_challenge_id_idx;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS next_recurrence_at,
  DROP COLUMN IF EXISTS recurrence_frequency,
  DROP COLUMN IF EXISTS recurring_challenge_id;

DROP INDEX IF EXISTS scenarios_next_recurrence_at_idx;
DROP INDEX IF EXISTS scenarios_recurring_scenario_id_idx;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS next_recurrence_at,
  DROP COLUMN IF EXISTS recurrence_frequency,
  DROP COLUMN IF EXISTS recurring_scenario_id;
