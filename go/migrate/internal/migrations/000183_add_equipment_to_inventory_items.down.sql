DROP TABLE IF EXISTS user_equipment;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS equip_slot,
  DROP COLUMN IF EXISTS strength_mod,
  DROP COLUMN IF EXISTS dexterity_mod,
  DROP COLUMN IF EXISTS constitution_mod,
  DROP COLUMN IF EXISTS intelligence_mod,
  DROP COLUMN IF EXISTS wisdom_mod,
  DROP COLUMN IF EXISTS charisma_mod;
