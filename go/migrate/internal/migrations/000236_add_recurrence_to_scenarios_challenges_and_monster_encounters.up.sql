ALTER TABLE scenarios
  ADD COLUMN recurring_scenario_id UUID,
  ADD COLUMN recurrence_frequency TEXT,
  ADD COLUMN next_recurrence_at TIMESTAMP;

CREATE INDEX scenarios_recurring_scenario_id_idx ON scenarios(recurring_scenario_id);
CREATE INDEX scenarios_next_recurrence_at_idx ON scenarios(next_recurrence_at);

ALTER TABLE challenges
  ADD COLUMN recurring_challenge_id UUID,
  ADD COLUMN recurrence_frequency TEXT,
  ADD COLUMN next_recurrence_at TIMESTAMP;

CREATE INDEX challenges_recurring_challenge_id_idx ON challenges(recurring_challenge_id);
CREATE INDEX challenges_next_recurrence_at_idx ON challenges(next_recurrence_at);

ALTER TABLE monster_encounters
  ADD COLUMN recurring_monster_encounter_id UUID,
  ADD COLUMN recurrence_frequency TEXT,
  ADD COLUMN next_recurrence_at TIMESTAMP;

CREATE INDEX monster_encounters_recurring_monster_encounter_id_idx ON monster_encounters(recurring_monster_encounter_id);
CREATE INDEX monster_encounters_next_recurrence_at_idx ON monster_encounters(next_recurrence_at);
