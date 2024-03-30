BEGIN;

ALTER TABLE how_many_questions DROP COLUMN prompt_seed_index;
ALTER TABLE how_many_questions DROP COLUMN prompt;

COMMIT;
