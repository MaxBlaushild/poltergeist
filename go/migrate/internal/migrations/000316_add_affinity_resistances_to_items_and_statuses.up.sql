ALTER TABLE inventory_items
    ADD COLUMN IF NOT EXISTS physical_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS piercing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS slashing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bludgeoning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fire_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ice_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lightning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS poison_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS arcane_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS holy_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shadow_resistance_percent INTEGER NOT NULL DEFAULT 0;

ALTER TABLE user_statuses
    ADD COLUMN IF NOT EXISTS physical_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS piercing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS slashing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bludgeoning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fire_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ice_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lightning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS poison_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS arcane_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS holy_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shadow_resistance_percent INTEGER NOT NULL DEFAULT 0;

ALTER TABLE monster_statuses
    ADD COLUMN IF NOT EXISTS physical_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS piercing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS slashing_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bludgeoning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fire_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ice_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lightning_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS poison_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS arcane_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS holy_resistance_percent INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shadow_resistance_percent INTEGER NOT NULL DEFAULT 0;
