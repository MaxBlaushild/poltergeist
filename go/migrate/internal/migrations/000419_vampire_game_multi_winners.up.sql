-- Games can now have multiple finishers per place. Move the single first/second/
-- third character columns to JSON arrays of character ids.
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS first_character_ids JSONB NOT NULL DEFAULT '[]';
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS second_character_ids JSONB NOT NULL DEFAULT '[]';
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS third_character_ids JSONB NOT NULL DEFAULT '[]';

UPDATE vampire_games SET first_character_ids = jsonb_build_array(first_character_id::text) WHERE first_character_id IS NOT NULL;
UPDATE vampire_games SET second_character_ids = jsonb_build_array(second_character_id::text) WHERE second_character_id IS NOT NULL;
UPDATE vampire_games SET third_character_ids = jsonb_build_array(third_character_id::text) WHERE third_character_id IS NOT NULL;

ALTER TABLE vampire_games DROP COLUMN IF EXISTS first_character_id;
ALTER TABLE vampire_games DROP COLUMN IF EXISTS second_character_id;
ALTER TABLE vampire_games DROP COLUMN IF EXISTS third_character_id;
