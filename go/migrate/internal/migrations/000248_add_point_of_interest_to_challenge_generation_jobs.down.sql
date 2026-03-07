BEGIN;

DROP INDEX IF EXISTS challenge_generation_jobs_point_of_interest_id_idx;

ALTER TABLE challenge_generation_jobs
  DROP COLUMN IF EXISTS point_of_interest_id;

COMMIT;
