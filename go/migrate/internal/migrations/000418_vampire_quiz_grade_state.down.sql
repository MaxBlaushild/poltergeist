ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS grade_status;
ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS grade_error;
ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS grade_started_at;
ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS grade_attempts;
