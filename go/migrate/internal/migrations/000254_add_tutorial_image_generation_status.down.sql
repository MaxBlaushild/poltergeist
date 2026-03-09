ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS image_generation_error,
  DROP COLUMN IF EXISTS image_generation_status;
