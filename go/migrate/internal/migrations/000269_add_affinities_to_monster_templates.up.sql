ALTER TABLE monster_templates
ADD COLUMN IF NOT EXISTS strong_against_affinity TEXT,
ADD COLUMN IF NOT EXISTS weak_against_affinity TEXT;
