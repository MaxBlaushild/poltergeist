BEGIN;

DELETE FROM base_structure_level_costs
WHERE structure_definition_id IN (
  SELECT id
  FROM base_structure_definitions
  WHERE key IN ('hearth', 'workshop', 'alchemy_lab', 'war_room')
);

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
SELECT id, 2, 'timber', 40 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 2, 'stone', 25 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 2, 'iron', 10 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'timber', 80 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'stone', 60 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'iron', 20 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 3, 'herbs', 10 FROM base_structure_definitions WHERE key = 'hearth'
UNION ALL
SELECT id, 1, 'timber', 60 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 1, 'stone', 35 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 1, 'iron', 20 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 2, 'iron', 40 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 2, 'arcane_dust', 20 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'iron', 60 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'arcane_dust', 30 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 3, 'relic_shards', 10 FROM base_structure_definitions WHERE key = 'workshop'
UNION ALL
SELECT id, 1, 'timber', 35 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'stone', 20 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'herbs', 30 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'arcane_dust', 10 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'herbs', 50 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'arcane_dust', 20 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 2, 'iron', 10 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'herbs', 80 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'arcane_dust', 35 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 3, 'relic_shards', 10 FROM base_structure_definitions WHERE key = 'alchemy_lab'
UNION ALL
SELECT id, 1, 'timber', 50 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 1, 'stone', 25 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 1, 'arcane_dust', 20 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 2, 'iron', 20 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 2, 'arcane_dust', 30 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'iron', 20 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'arcane_dust', 50 FROM base_structure_definitions WHERE key = 'war_room'
UNION ALL
SELECT id, 3, 'relic_shards', 12 FROM base_structure_definitions WHERE key = 'war_room'
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;

COMMIT;
