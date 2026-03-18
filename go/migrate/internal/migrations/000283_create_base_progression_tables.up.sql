BEGIN;

CREATE TABLE IF NOT EXISTS base_resource_balances (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  resource_key TEXT NOT NULL,
  amount INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, resource_key)
);

CREATE TABLE IF NOT EXISTS base_resource_ledger (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  resource_key TEXT NOT NULL,
  delta INTEGER NOT NULL,
  source_type TEXT NOT NULL,
  source_id UUID,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS base_structure_definitions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  key TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL,
  max_level INTEGER NOT NULL DEFAULT 1,
  sort_order INTEGER NOT NULL DEFAULT 0,
  image_url TEXT NOT NULL DEFAULT '',
  effect_type TEXT NOT NULL,
  effect_config JSONB NOT NULL DEFAULT '{}'::jsonb,
  prereq_config JSONB NOT NULL DEFAULT '{}'::jsonb,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS base_structure_level_costs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  structure_definition_id UUID NOT NULL REFERENCES base_structure_definitions(id) ON DELETE CASCADE,
  level INTEGER NOT NULL,
  resource_key TEXT NOT NULL,
  amount INTEGER NOT NULL,
  UNIQUE (structure_definition_id, level, resource_key)
);

CREATE TABLE IF NOT EXISTS user_base_structures (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  structure_key TEXT NOT NULL REFERENCES base_structure_definitions(key) ON DELETE CASCADE,
  level INTEGER NOT NULL DEFAULT 1,
  UNIQUE (base_id, structure_key)
);

CREATE TABLE IF NOT EXISTS user_base_daily_state (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  state_key TEXT NOT NULL,
  state_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  resets_on DATE NOT NULL,
  UNIQUE (user_id, state_key, resets_on)
);

CREATE INDEX IF NOT EXISTS idx_base_resource_ledger_user_created_at
  ON base_resource_ledger(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_base_structure_definitions_active_sort
  ON base_structure_definitions(active, sort_order ASC, name ASC);

CREATE INDEX IF NOT EXISTS idx_base_structure_level_costs_definition_level
  ON base_structure_level_costs(structure_definition_id, level ASC);

CREATE INDEX IF NOT EXISTS idx_user_base_structures_user_id
  ON user_base_structures(user_id);

CREATE INDEX IF NOT EXISTS idx_user_base_daily_state_user_resets_on
  ON user_base_daily_state(user_id, resets_on DESC);

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
    'hearth',
    'Hearth',
    'The center of your base. Rest here to recover and prepare for the road ahead.',
    'core',
    3,
    10,
    'rest_bonus',
    '{
      "valuesByLevel": {
        "1": { "restoresHealth": true, "restoresMana": true },
        "2": { "restoresHealth": true, "restoresMana": true, "removeNegativeStatuses": 1 },
        "3": {
          "restoresHealth": true,
          "restoresMana": true,
          "removeNegativeStatuses": 1,
          "grantsBuff": { "key": "rested", "healthPercent": 5, "manaPercent": 5, "durationEncounters": 2 }
        }
      }
    }'::jsonb,
    '{}'::jsonb
  ),
  (
    'workshop',
    'Workshop',
    'A place to refine tools and prepare practical gear for expeditions.',
    'utility',
    3,
    20,
    'craft_unlock',
    '{
      "valuesByLevel": {
        "1": { "recipes": ["lock_kit_basic", "field_rations"] },
        "2": { "recipes": ["lock_kit_basic", "field_rations"], "materialCostReductionPercent": 10 },
        "3": { "recipes": ["lock_kit_basic", "field_rations", "scouting_tools"], "freeUtilityCraftsPerDay": 1 }
      }
    }'::jsonb,
    '{ "requiredStructures": [{ "key": "hearth", "level": 1 }] }'::jsonb
  ),
  (
    'alchemy_lab',
    'Alchemy Lab',
    'Distill gathered herbs and essence into field tonics and restorative brews.',
    'crafting',
    3,
    30,
    'daily_choice_buff',
    '{
      "actionKey": "alchemy_tonic",
      "valuesByLevel": {
        "1": { "choices": ["healing_tonic", "mana_tonic", "antidote_tonic"], "durationEncounters": 2 },
        "2": { "choices": ["healing_tonic", "mana_tonic", "antidote_tonic"], "durationEncounters": 3, "dailyIngredientCache": true },
        "3": { "choices": ["healing_tonic", "mana_tonic", "antidote_tonic"], "durationEncounters": 3, "dailyIngredientCache": true, "bonusPotencyPercent": 10 }
      }
    }'::jsonb,
    '{ "requiredStructures": [{ "key": "hearth", "level": 1 }] }'::jsonb
  ),
  (
    'war_room',
    'War Room',
    'Maps, notes, and hard-won intelligence gathered from the wider world.',
    'intel',
    3,
    40,
    'reward_bias',
    '{
      "actionKey": "zone_focus",
      "valuesByLevel": {
        "1": { "trackedQuestHintStrength": "improved" },
        "2": { "trackedQuestHintStrength": "strong", "revealsNearbyActivityClusters": 1 },
        "3": {
          "trackedQuestHintStrength": "strong",
          "revealsNearbyActivityClusters": 1,
          "focusChoices": ["monster_hunting", "questing", "scavenging"],
          "rewardBiasPercent": 10
        }
      }
    }'::jsonb,
    '{ "requiredStructures": [{ "key": "hearth", "level": 1 }] }'::jsonb
  )
ON CONFLICT (key) DO NOTHING;

INSERT INTO user_base_structures (id, created_at, updated_at, base_id, user_id, structure_key, level)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  b.id,
  b.user_id,
  'hearth',
  1
FROM bases b
LEFT JOIN user_base_structures ubs
  ON ubs.base_id = b.id
  AND ubs.structure_key = 'hearth'
WHERE ubs.id IS NULL
ON CONFLICT (base_id, structure_key) DO NOTHING;

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
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
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
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
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
ON CONFLICT (structure_definition_id, level, resource_key) DO NOTHING;

INSERT INTO base_structure_level_costs (structure_definition_id, level, resource_key, amount)
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
