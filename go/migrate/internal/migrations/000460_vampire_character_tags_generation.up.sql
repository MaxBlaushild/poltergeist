-- Tracks the AI tag-generation job's state per character, mirroring
-- vampire_quiz_submissions' grade_status/grade_error columns (same
-- queued/generating/generated/failed lifecycle, same "poll and show the
-- error" UI pattern).
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS tags_generation_status TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS tags_generation_error TEXT NOT NULL DEFAULT '';
