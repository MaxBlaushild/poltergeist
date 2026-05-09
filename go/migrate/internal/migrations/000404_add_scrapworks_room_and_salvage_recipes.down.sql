DELETE FROM user_base_structures
WHERE structure_key = 'scrapworks';

DELETE FROM base_structure_level_costs
WHERE structure_definition_id IN (
  SELECT id FROM base_structure_definitions WHERE key = 'scrapworks'
);

DELETE FROM base_structure_definitions
WHERE key = 'scrapworks';

ALTER TABLE inventory_items
DROP COLUMN IF EXISTS scrapworks_recipes;
