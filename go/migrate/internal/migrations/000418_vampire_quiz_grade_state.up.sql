-- Per-submission grading state machine: '' (never) → queued → grading → graded | failed.
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS grade_status TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS grade_error TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS grade_started_at TIMESTAMPTZ;
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS grade_attempts INTEGER NOT NULL DEFAULT 0;
