-- A one-line AI rationale for each Part 1 open-end score, shown to GMs alongside
-- the recommended Blood Token value.
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS ai_rationale TEXT NOT NULL DEFAULT '';
