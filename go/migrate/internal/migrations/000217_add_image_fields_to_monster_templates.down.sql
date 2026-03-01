ALTER TABLE monster_templates
  DROP COLUMN IF EXISTS thumbnail_url;

ALTER TABLE monster_templates
  DROP COLUMN IF EXISTS image_url;
