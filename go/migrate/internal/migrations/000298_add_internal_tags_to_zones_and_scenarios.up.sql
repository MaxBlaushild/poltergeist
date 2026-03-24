ALTER TABLE zones
  ADD COLUMN IF NOT EXISTS internal_tags JSONB;

UPDATE zones
SET internal_tags = '[]'::jsonb
WHERE internal_tags IS NULL;

ALTER TABLE zones
  ALTER COLUMN internal_tags SET DEFAULT '[]'::jsonb,
  ALTER COLUMN internal_tags SET NOT NULL;

ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS internal_tags JSONB;

UPDATE scenarios
SET internal_tags = '[]'::jsonb
WHERE internal_tags IS NULL;

ALTER TABLE scenarios
  ALTER COLUMN internal_tags SET DEFAULT '[]'::jsonb,
  ALTER COLUMN internal_tags SET NOT NULL;
