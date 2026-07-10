-- Revert to a single finisher per place (keeps the first id from each array).
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS first_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL;
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS second_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL;
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS third_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL;

UPDATE vampire_games SET first_character_id = (first_character_ids->>0)::uuid WHERE jsonb_array_length(first_character_ids) > 0;
UPDATE vampire_games SET second_character_id = (second_character_ids->>0)::uuid WHERE jsonb_array_length(second_character_ids) > 0;
UPDATE vampire_games SET third_character_id = (third_character_ids->>0)::uuid WHERE jsonb_array_length(third_character_ids) > 0;

ALTER TABLE vampire_games DROP COLUMN IF EXISTS first_character_ids;
ALTER TABLE vampire_games DROP COLUMN IF EXISTS second_character_ids;
ALTER TABLE vampire_games DROP COLUMN IF EXISTS third_character_ids;
