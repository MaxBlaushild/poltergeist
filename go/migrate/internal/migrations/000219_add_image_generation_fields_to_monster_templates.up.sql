ALTER TABLE monster_templates
  ADD COLUMN IF NOT EXISTS image_generation_status TEXT NOT NULL DEFAULT 'none',
  ADD COLUMN IF NOT EXISTS image_generation_error TEXT;
