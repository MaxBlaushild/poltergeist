ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS spell_damage_bonus_percent,
  DROP COLUMN IF EXISTS damage_blocked,
  DROP COLUMN IF EXISTS block_percentage,
  DROP COLUMN IF EXISTS swipes_per_attack,
  DROP COLUMN IF EXISTS damage_max,
  DROP COLUMN IF EXISTS damage_min,
  DROP COLUMN IF EXISTS handedness,
  DROP COLUMN IF EXISTS hand_item_category;
