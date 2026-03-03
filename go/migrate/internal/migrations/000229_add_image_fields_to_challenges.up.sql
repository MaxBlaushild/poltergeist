ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS thumbnail_url TEXT NOT NULL DEFAULT '';

UPDATE challenges
SET thumbnail_url = image_url
WHERE COALESCE(thumbnail_url, '') = '' AND COALESCE(image_url, '') <> '';
