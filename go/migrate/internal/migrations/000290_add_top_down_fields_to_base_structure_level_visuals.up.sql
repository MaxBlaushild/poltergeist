ALTER TABLE base_structure_level_visuals
ADD COLUMN IF NOT EXISTS top_down_image_url TEXT NOT NULL DEFAULT '';

ALTER TABLE base_structure_level_visuals
ADD COLUMN IF NOT EXISTS top_down_thumbnail_url TEXT NOT NULL DEFAULT '';

ALTER TABLE base_structure_level_visuals
ADD COLUMN IF NOT EXISTS top_down_image_generation_status TEXT NOT NULL DEFAULT 'none';

ALTER TABLE base_structure_level_visuals
ADD COLUMN IF NOT EXISTS top_down_image_generation_error TEXT;
