ALTER TABLE characters
  ADD COLUMN image_generation_status TEXT NOT NULL DEFAULT 'none',
  ADD COLUMN image_generation_error TEXT;
