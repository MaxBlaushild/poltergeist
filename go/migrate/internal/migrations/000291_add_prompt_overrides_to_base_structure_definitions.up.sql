ALTER TABLE base_structure_definitions
ADD COLUMN IF NOT EXISTS image_prompt TEXT NOT NULL DEFAULT '';

ALTER TABLE base_structure_definitions
ADD COLUMN IF NOT EXISTS top_down_image_prompt TEXT NOT NULL DEFAULT '';
