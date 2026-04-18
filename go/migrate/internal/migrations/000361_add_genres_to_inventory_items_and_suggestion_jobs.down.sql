BEGIN;

DROP INDEX IF EXISTS idx_inventory_item_suggestion_jobs_genre_id;
DROP INDEX IF EXISTS idx_inventory_items_genre_id;

ALTER TABLE inventory_item_suggestion_jobs
  DROP CONSTRAINT IF EXISTS inventory_item_suggestion_jobs_genre_id_fkey;

ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_genre_id_fkey;

ALTER TABLE inventory_item_suggestion_jobs
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS genre_id;

COMMIT;
