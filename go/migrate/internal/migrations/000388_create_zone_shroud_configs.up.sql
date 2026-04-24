CREATE TABLE IF NOT EXISTS zone_shroud_configs (
  id INTEGER PRIMARY KEY,
  pattern_tile_url TEXT NOT NULL DEFAULT '',
  pattern_tile_prompt TEXT NOT NULL DEFAULT '',
  pattern_tile_generation_status TEXT NOT NULL DEFAULT 'none',
  pattern_tile_generation_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO zone_shroud_configs (
  id,
  pattern_tile_url,
  pattern_tile_prompt,
  pattern_tile_generation_status,
  pattern_tile_generation_error
)
VALUES (1, '', '', 'none', '')
ON CONFLICT (id) DO NOTHING;
