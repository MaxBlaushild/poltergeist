ALTER TABLE monster_templates
  DROP COLUMN IF EXISTS image_generation_error,
  DROP COLUMN IF EXISTS image_generation_status;
