ALTER TABLE monster_templates
DROP COLUMN IF EXISTS strong_against_affinity,
DROP COLUMN IF EXISTS weak_against_affinity;
