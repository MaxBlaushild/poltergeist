ALTER TABLE inventory_items
  ADD COLUMN physical_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN piercing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN slashing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN bludgeoning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN fire_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN ice_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN lightning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN poison_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN arcane_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN holy_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN shadow_damage_bonus_percent INTEGER NOT NULL DEFAULT 0;

ALTER TABLE user_statuses
  ADD COLUMN physical_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN piercing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN slashing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN bludgeoning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN fire_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN ice_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN lightning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN poison_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN arcane_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN holy_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN shadow_damage_bonus_percent INTEGER NOT NULL DEFAULT 0;

ALTER TABLE monster_statuses
  ADD COLUMN physical_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN piercing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN slashing_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN bludgeoning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN fire_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN ice_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN lightning_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN poison_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN arcane_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN holy_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN shadow_damage_bonus_percent INTEGER NOT NULL DEFAULT 0;
