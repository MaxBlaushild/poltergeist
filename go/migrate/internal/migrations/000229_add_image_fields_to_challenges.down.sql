ALTER TABLE challenges
  DROP COLUMN IF EXISTS thumbnail_url,
  DROP COLUMN IF EXISTS image_url;
