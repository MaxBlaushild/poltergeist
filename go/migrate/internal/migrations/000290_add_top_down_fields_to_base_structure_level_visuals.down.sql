ALTER TABLE base_structure_level_visuals
DROP COLUMN IF EXISTS top_down_image_generation_error;

ALTER TABLE base_structure_level_visuals
DROP COLUMN IF EXISTS top_down_image_generation_status;

ALTER TABLE base_structure_level_visuals
DROP COLUMN IF EXISTS top_down_thumbnail_url;

ALTER TABLE base_structure_level_visuals
DROP COLUMN IF EXISTS top_down_image_url;
