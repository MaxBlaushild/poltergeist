ALTER TABLE monster_templates
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
  ADD COLUMN shadow_damage_bonus_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN physical_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN piercing_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN slashing_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN bludgeoning_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN fire_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN ice_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN lightning_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN poison_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN arcane_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN holy_resistance_percent INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN shadow_resistance_percent INTEGER NOT NULL DEFAULT 0;

UPDATE monster_templates SET physical_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'physical';
UPDATE monster_templates SET piercing_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'piercing';
UPDATE monster_templates SET slashing_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'slashing';
UPDATE monster_templates SET bludgeoning_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'bludgeoning';
UPDATE monster_templates SET fire_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'fire';
UPDATE monster_templates SET ice_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'ice';
UPDATE monster_templates SET lightning_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'lightning';
UPDATE monster_templates SET poison_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'poison';
UPDATE monster_templates SET arcane_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'arcane';
UPDATE monster_templates SET holy_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'holy';
UPDATE monster_templates SET shadow_resistance_percent = 50 WHERE LOWER(TRIM(COALESCE(strong_against_affinity, ''))) = 'shadow';

UPDATE monster_templates SET physical_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'physical';
UPDATE monster_templates SET piercing_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'piercing';
UPDATE monster_templates SET slashing_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'slashing';
UPDATE monster_templates SET bludgeoning_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'bludgeoning';
UPDATE monster_templates SET fire_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'fire';
UPDATE monster_templates SET ice_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'ice';
UPDATE monster_templates SET lightning_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'lightning';
UPDATE monster_templates SET poison_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'poison';
UPDATE monster_templates SET arcane_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'arcane';
UPDATE monster_templates SET holy_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'holy';
UPDATE monster_templates SET shadow_resistance_percent = -100 WHERE LOWER(TRIM(COALESCE(weak_against_affinity, ''))) = 'shadow';
