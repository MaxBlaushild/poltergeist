BEGIN;

DROP INDEX IF EXISTS idx_point_of_interest_imports_genre_id;
DROP INDEX IF EXISTS idx_points_of_interest_genre_id;

ALTER TABLE point_of_interest_imports
  DROP CONSTRAINT IF EXISTS point_of_interest_imports_genre_id_fkey;

ALTER TABLE points_of_interest
  DROP CONSTRAINT IF EXISTS points_of_interest_genre_id_fkey;

ALTER TABLE point_of_interest_imports
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE points_of_interest
  DROP COLUMN IF EXISTS genre_id;

COMMIT;
