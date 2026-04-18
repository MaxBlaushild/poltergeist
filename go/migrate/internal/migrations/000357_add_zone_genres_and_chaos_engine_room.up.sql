CREATE TABLE IF NOT EXISTS zone_genres (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_zone_genres_name_lower_unique
  ON zone_genres ((LOWER(name)));

CREATE INDEX IF NOT EXISTS idx_zone_genres_active_sort
  ON zone_genres (active, sort_order ASC, name ASC);

CREATE TABLE IF NOT EXISTS zone_genre_scores (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  genre_id UUID NOT NULL REFERENCES zone_genres(id) ON DELETE CASCADE,
  score INTEGER NOT NULL DEFAULT 0,
  CONSTRAINT chk_zone_genre_scores_nonnegative CHECK (score >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_zone_genre_scores_zone_genre_unique
  ON zone_genre_scores (zone_id, genre_id);

CREATE INDEX IF NOT EXISTS idx_zone_genre_scores_genre
  ON zone_genre_scores (genre_id, score DESC, zone_id);

INSERT INTO base_structure_definitions (
  key,
  name,
  description,
  category,
  max_level,
  sort_order,
  effect_type,
  effect_config,
  prereq_config
)
VALUES (
  'chaos_engine',
  'Chaos Engine Room',
  'Feed a configured catalyst into the engine to shift a zone one point toward a chosen story genre for the whole world.',
  'attunement',
  1,
  50,
  'zone_genre',
  '{ "requiredInventoryItemId": null }'::jsonb,
  '{ "requiredStructures": [{ "key": "hearth", "level": 1 }] }'::jsonb
)
ON CONFLICT (key) DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 1, 'timber', 18 FROM base_structure_definitions WHERE key = 'chaos_engine'
ON CONFLICT DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 1, 'stone', 12 FROM base_structure_definitions WHERE key = 'chaos_engine'
ON CONFLICT DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 1, 'arcane_dust', 8 FROM base_structure_definitions WHERE key = 'chaos_engine'
ON CONFLICT DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 1, 'relic_shards', 2 FROM base_structure_definitions WHERE key = 'chaos_engine'
ON CONFLICT DO NOTHING;
