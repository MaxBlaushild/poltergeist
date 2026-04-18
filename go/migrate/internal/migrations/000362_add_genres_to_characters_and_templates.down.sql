BEGIN;

DROP INDEX IF EXISTS idx_character_templates_genre_id;
DROP INDEX IF EXISTS idx_characters_genre_id;

ALTER TABLE character_templates
  DROP CONSTRAINT IF EXISTS character_templates_genre_id_fkey;

ALTER TABLE characters
  DROP CONSTRAINT IF EXISTS characters_genre_id_fkey;

ALTER TABLE character_templates
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE characters
  DROP COLUMN IF EXISTS genre_id;

COMMIT;
