ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS loadout_objective_copy TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS base_kit_objective_copy TEXT NOT NULL DEFAULT '';

UPDATE tutorial_configs
SET loadout_objective_copy = 'Equip your new gear and use the spellbook to continue.'
WHERE TRIM(COALESCE(loadout_objective_copy, '')) = '';

UPDATE tutorial_configs
SET base_kit_objective_copy = 'Use your home base kit to establish your base.'
WHERE TRIM(COALESCE(base_kit_objective_copy, '')) = '';
