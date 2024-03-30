BEGIN;

ALTER TABLE how_many_questions ADD COLUMN prompt_seed_index int;
UPDATE how_many_questions SET prompt_seed_index = 0;
ALTER TABLE how_many_questions ADD COLUMN prompt TEXT;

COMMIT;
