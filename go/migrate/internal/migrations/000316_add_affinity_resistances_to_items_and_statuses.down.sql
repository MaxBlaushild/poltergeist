ALTER TABLE monster_statuses
    DROP COLUMN IF EXISTS shadow_resistance_percent,
    DROP COLUMN IF EXISTS holy_resistance_percent,
    DROP COLUMN IF EXISTS arcane_resistance_percent,
    DROP COLUMN IF EXISTS poison_resistance_percent,
    DROP COLUMN IF EXISTS lightning_resistance_percent,
    DROP COLUMN IF EXISTS ice_resistance_percent,
    DROP COLUMN IF EXISTS fire_resistance_percent,
    DROP COLUMN IF EXISTS bludgeoning_resistance_percent,
    DROP COLUMN IF EXISTS slashing_resistance_percent,
    DROP COLUMN IF EXISTS piercing_resistance_percent,
    DROP COLUMN IF EXISTS physical_resistance_percent;

ALTER TABLE user_statuses
    DROP COLUMN IF EXISTS shadow_resistance_percent,
    DROP COLUMN IF EXISTS holy_resistance_percent,
    DROP COLUMN IF EXISTS arcane_resistance_percent,
    DROP COLUMN IF EXISTS poison_resistance_percent,
    DROP COLUMN IF EXISTS lightning_resistance_percent,
    DROP COLUMN IF EXISTS ice_resistance_percent,
    DROP COLUMN IF EXISTS fire_resistance_percent,
    DROP COLUMN IF EXISTS bludgeoning_resistance_percent,
    DROP COLUMN IF EXISTS slashing_resistance_percent,
    DROP COLUMN IF EXISTS piercing_resistance_percent,
    DROP COLUMN IF EXISTS physical_resistance_percent;

ALTER TABLE inventory_items
    DROP COLUMN IF EXISTS shadow_resistance_percent,
    DROP COLUMN IF EXISTS holy_resistance_percent,
    DROP COLUMN IF EXISTS arcane_resistance_percent,
    DROP COLUMN IF EXISTS poison_resistance_percent,
    DROP COLUMN IF EXISTS lightning_resistance_percent,
    DROP COLUMN IF EXISTS ice_resistance_percent,
    DROP COLUMN IF EXISTS fire_resistance_percent,
    DROP COLUMN IF EXISTS bludgeoning_resistance_percent,
    DROP COLUMN IF EXISTS slashing_resistance_percent,
    DROP COLUMN IF EXISTS piercing_resistance_percent,
    DROP COLUMN IF EXISTS physical_resistance_percent;
