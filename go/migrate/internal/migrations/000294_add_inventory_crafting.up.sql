ALTER TABLE inventory_items
ADD COLUMN IF NOT EXISTS consume_teach_recipe_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS alchemy_recipes JSONB NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS workshop_recipes JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE IF NOT EXISTS user_learned_recipes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  recipe_id TEXT NOT NULL,
  learned_from_inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE SET NULL,
  learned_from_owned_item_id UUID REFERENCES owned_inventory_items(id) ON DELETE SET NULL,
  UNIQUE (user_id, recipe_id)
);

CREATE INDEX IF NOT EXISTS idx_user_learned_recipes_user_id
  ON user_learned_recipes(user_id);
