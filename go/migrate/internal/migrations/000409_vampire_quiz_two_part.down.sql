ALTER TABLE vampire_missions DROP COLUMN IF EXISTS sabotage_hf;
ALTER TABLE vampire_missions DROP COLUMN IF EXISTS sabotage_house_id;

ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS awarded_bt;
ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS ai_score;

ALTER TABLE vampire_quiz_questions ADD COLUMN IF NOT EXISTS hf_effect JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS tier;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS hf_value;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS max_bt;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS rubric;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS part;

ALTER TABLE vampire_house_favor_ledger ALTER COLUMN delta TYPE INTEGER USING round(delta)::integer;

ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS quiz_open BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS quiz_part1_opened_at;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS quiz_part2_open;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS quiz_part1_open;
