ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE movement_patterns
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE resources
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE healing_fountains
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE treasure_chests
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE monsters
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE expositions
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE quests
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE points_of_interest
  DROP COLUMN IF EXISTS zone_kind;
