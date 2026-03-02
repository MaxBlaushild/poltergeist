ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS damage_affinity TEXT;

UPDATE inventory_items
SET damage_affinity = 'physical'
WHERE damage_affinity IS NULL
  AND damage_min IS NOT NULL
  AND damage_max IS NOT NULL;
