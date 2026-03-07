BEGIN;

ALTER TABLE challenge_generation_jobs
  ADD COLUMN IF NOT EXISTS point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS challenge_generation_jobs_point_of_interest_id_idx
  ON challenge_generation_jobs(point_of_interest_id);

COMMIT;
