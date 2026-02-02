BEGIN;

DROP INDEX IF EXISTS idx_characters_point_of_interest_id;
ALTER TABLE characters DROP COLUMN IF EXISTS point_of_interest_id;

COMMIT;
