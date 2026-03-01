ALTER TABLE monster_templates
  ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

ALTER TABLE monster_templates
  ADD COLUMN IF NOT EXISTS thumbnail_url TEXT NOT NULL DEFAULT '';

UPDATE monster_templates
SET thumbnail_url = image_url
WHERE COALESCE(thumbnail_url, '') = '' AND COALESCE(image_url, '') <> '';
