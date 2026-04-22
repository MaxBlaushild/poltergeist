ALTER TABLE zone_kinds
  DROP COLUMN IF EXISTS pattern_tile_generation_error,
  DROP COLUMN IF EXISTS pattern_tile_generation_status,
  DROP COLUMN IF EXISTS pattern_tile_prompt,
  DROP COLUMN IF EXISTS pattern_tile_url;
