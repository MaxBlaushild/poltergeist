ALTER TABLE scenarios
  DROP COLUMN IF EXISTS internal_tags;

ALTER TABLE zones
  DROP COLUMN IF EXISTS internal_tags;
