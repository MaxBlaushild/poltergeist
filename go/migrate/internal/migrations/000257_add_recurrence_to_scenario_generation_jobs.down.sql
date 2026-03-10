ALTER TABLE scenario_generation_jobs
  DROP COLUMN IF EXISTS next_recurrence_at,
  DROP COLUMN IF EXISTS recurrence_frequency,
  DROP COLUMN IF EXISTS recurring_scenario_id;
