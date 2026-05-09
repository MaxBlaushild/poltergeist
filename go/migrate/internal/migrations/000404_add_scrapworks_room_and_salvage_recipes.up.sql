ALTER TABLE inventory_items
ADD COLUMN IF NOT EXISTS scrapworks_recipes JSONB NOT NULL DEFAULT '[]'::jsonb;

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
VALUES
  (
    'scrapworks',
    'Scrapworks',
    'Break down worn gear and surplus curios into reusable components.',
    'utility',
    3,
    25,
    'salvage_unlock',
    '{
      "actionKey": "salvage",
      "valuesByLevel": {
        "1": { "recipeTier": 1 },
        "2": { "recipeTier": 2 },
        "3": { "recipeTier": 3 }
      }
    }'::jsonb,
    '{ "requiredStructures": [{ "key": "hearth", "level": 1 }] }'::jsonb
  )
ON CONFLICT (key) DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 1, 'timber', 10 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 1, 'stone', 8 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 1, 'iron', 6 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 2, 'stone', 8 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 2, 'iron', 10 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 2, 'monster_parts', 6 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 3, 'iron', 14 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 3, 'arcane_dust', 8 FROM base_structure_definitions WHERE key = 'scrapworks'
UNION ALL
SELECT id, 3, 'relic_shards', 3 FROM base_structure_definitions WHERE key = 'scrapworks'
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;
