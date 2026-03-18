CREATE TABLE IF NOT EXISTS base_structure_level_visuals (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  structure_definition_id UUID NOT NULL REFERENCES base_structure_definitions(id) ON DELETE CASCADE,
  level INTEGER NOT NULL,
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  image_generation_status TEXT NOT NULL DEFAULT 'none',
  image_generation_error TEXT,
  UNIQUE (structure_definition_id, level)
);

INSERT INTO base_structure_level_visuals (
  structure_definition_id,
  level,
  image_generation_status
)
SELECT
  d.id,
  levels.level,
  'none'
FROM base_structure_definitions d
CROSS JOIN LATERAL generate_series(1, GREATEST(d.max_level, 1)) AS levels(level)
ON CONFLICT (structure_definition_id, level) DO NOTHING;
