BEGIN;

DROP INDEX IF EXISTS idx_spells_genre_id;

ALTER TABLE spells
  DROP CONSTRAINT IF EXISTS spells_genre_id_fkey;

ALTER TABLE spells
  DROP COLUMN IF EXISTS genre_id;

COMMIT;
