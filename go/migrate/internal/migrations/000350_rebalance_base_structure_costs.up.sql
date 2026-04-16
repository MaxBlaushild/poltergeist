BEGIN;

DELETE FROM base_structure_level_costs
WHERE structure_definition_id IN (
  SELECT id
  FROM base_structure_definitions
  WHERE key IN ('hearth', 'workshop', 'alchemy_lab', 'war_room')
);

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 2, 'timber', 18 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 2, 'stone', 12 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 2, 'iron', 4 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'timber', 32 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'stone', 22 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'iron', 10 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'herbs', 6 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 1, 'timber', 16 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 1, 'stone', 10 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 2, 'timber', 8 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 2, 'stone', 6 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 2, 'iron', 8 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'timber', 12 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'iron', 16 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'arcane_dust', 8 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'relic_shards', 4 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 1, 'timber', 8 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'stone', 4 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'herbs', 14 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'stone', 6 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'herbs', 18 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'arcane_dust', 6 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'herbs', 28 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'arcane_dust', 12 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'iron', 6 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'relic_shards', 4 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'timber', 12 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 1, 'stone', 8 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 2, 'stone', 8 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 2, 'iron', 6 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 2, 'arcane_dust', 8 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'iron', 10 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'arcane_dust', 18 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'relic_shards', 5 FROM base_structure_definitions WHERE key = 'war_room'
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;

COMMIT;
