ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS image_generation_status TEXT NOT NULL DEFAULT 'none',
  ADD COLUMN IF NOT EXISTS image_generation_error TEXT;

UPDATE tutorial_configs
SET image_generation_status = 'none'
WHERE COALESCE(TRIM(image_generation_status), '') = '';
