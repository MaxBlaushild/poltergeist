ALTER TABLE zone_kinds
  ADD COLUMN pattern_tile_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN pattern_tile_prompt TEXT NOT NULL DEFAULT '',
  ADD COLUMN pattern_tile_generation_status TEXT NOT NULL DEFAULT 'none',
  ADD COLUMN pattern_tile_generation_error TEXT NOT NULL DEFAULT '';
