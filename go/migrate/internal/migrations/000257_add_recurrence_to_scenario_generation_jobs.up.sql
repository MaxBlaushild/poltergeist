ALTER TABLE scenario_generation_jobs
  ADD COLUMN IF NOT EXISTS recurring_scenario_id UUID,
  ADD COLUMN IF NOT EXISTS recurrence_frequency TEXT,
  ADD COLUMN IF NOT EXISTS next_recurrence_at TIMESTAMP WITH TIME ZONE;
