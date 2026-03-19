BEGIN;

ALTER TABLE user_base_structures
ADD COLUMN IF NOT EXISTS grid_x INTEGER;

ALTER TABLE user_base_structures
ADD COLUMN IF NOT EXISTS grid_y INTEGER;

UPDATE user_base_structures
SET
  grid_x = CASE structure_key
    WHEN 'hearth' THEN 2
    WHEN 'workshop' THEN 2
    WHEN 'alchemy_lab' THEN 3
    WHEN 'war_room' THEN 2
    ELSE 1
  END,
  grid_y = CASE structure_key
    WHEN 'hearth' THEN 2
    WHEN 'workshop' THEN 1
    WHEN 'alchemy_lab' THEN 2
    WHEN 'war_room' THEN 3
    ELSE 1
  END
WHERE grid_x IS NULL OR grid_y IS NULL;

ALTER TABLE user_base_structures
ALTER COLUMN grid_x SET NOT NULL;

ALTER TABLE user_base_structures
ALTER COLUMN grid_y SET NOT NULL;

ALTER TABLE user_base_structures
ADD CONSTRAINT user_base_structures_base_id_grid_unique UNIQUE (base_id, grid_x, grid_y);

COMMIT;
