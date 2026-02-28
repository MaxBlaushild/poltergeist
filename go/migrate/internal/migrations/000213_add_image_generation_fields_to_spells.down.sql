ALTER TABLE spells
  DROP COLUMN IF EXISTS image_generation_error,
  DROP COLUMN IF EXISTS image_generation_status;
