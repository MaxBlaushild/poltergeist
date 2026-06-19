-- Two-part end quiz, decimal House Favor, and sabotage-mission HF deductions.

-- Game state: replace the single quiz flag with two-part flags + Part 1 timer start.
ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS quiz_part1_open BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS quiz_part2_open BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS quiz_part1_opened_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS quiz_open;

-- House Favor must support fractional values from Part 2's normalized scoring.
ALTER TABLE vampire_house_favor_ledger ALTER COLUMN delta TYPE NUMERIC USING delta::numeric;

-- Quiz questions: two-part model.
--   part 1 -> open-end, AI-graded (rubric + max_bt)
--   part 2 -> multiple choice, normalized HF (hf_value + tier)
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS part INTEGER NOT NULL DEFAULT 2;
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS rubric TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS max_bt INTEGER NOT NULL DEFAULT 0;
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS hf_value NUMERIC NOT NULL DEFAULT 0;
ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS hf_effect;

-- Quiz submissions: Part 1 AI score + the Blood Tokens awarded for it.
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS ai_score NUMERIC;
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS awarded_bt INTEGER NOT NULL DEFAULT 0;

-- Missions: a rare "sabotage" mission deducts House Favor from a target house
-- when the GM verifies it.
ALTER TABLE vampire_missions ADD COLUMN IF NOT EXISTS sabotage_house_id UUID REFERENCES vampire_houses(id) ON DELETE SET NULL;
ALTER TABLE vampire_missions ADD COLUMN IF NOT EXISTS sabotage_hf INTEGER NOT NULL DEFAULT 0;
