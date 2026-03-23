DROP TABLE IF EXISTS user_learned_recipes;

ALTER TABLE inventory_items
DROP COLUMN IF EXISTS workshop_recipes,
DROP COLUMN IF EXISTS alchemy_recipes,
DROP COLUMN IF EXISTS consume_teach_recipe_ids;
