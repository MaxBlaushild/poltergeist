BEGIN;

ALTER TABLE user_base_structures
DROP CONSTRAINT IF EXISTS user_base_structures_base_id_grid_unique;

ALTER TABLE user_base_structures
DROP COLUMN IF EXISTS grid_y;

ALTER TABLE user_base_structures
DROP COLUMN IF EXISTS grid_x;

COMMIT;
