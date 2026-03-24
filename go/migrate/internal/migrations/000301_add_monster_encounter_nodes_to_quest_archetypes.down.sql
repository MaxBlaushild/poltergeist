DELETE FROM quest_archetype_node_challenges
WHERE quest_archetype_node_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'monster_encounter'
)
OR quest_archetype_challenge_id IN (
  SELECT id
  FROM quest_archetype_challenges
  WHERE unlocked_node_id IN (
    SELECT id FROM quest_archetype_nodes WHERE node_type = 'monster_encounter'
  )
);

DELETE FROM quest_archetype_challenges
WHERE unlocked_node_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'monster_encounter'
);

DELETE FROM quest_archetypes
WHERE root_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'monster_encounter'
);

DELETE FROM quest_archetype_nodes
WHERE node_type = 'monster_encounter';

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS encounter_proximity_meters,
  DROP COLUMN IF EXISTS encounter_item_rewards_json,
  DROP COLUMN IF EXISTS encounter_material_rewards_json,
  DROP COLUMN IF EXISTS encounter_reward_gold,
  DROP COLUMN IF EXISTS encounter_reward_experience,
  DROP COLUMN IF EXISTS encounter_random_reward_size,
  DROP COLUMN IF EXISTS encounter_reward_mode,
  DROP COLUMN IF EXISTS target_level,
  DROP COLUMN IF EXISTS monster_ids,
  DROP COLUMN IF EXISTS node_type;

ALTER TABLE quest_archetype_nodes
  ALTER COLUMN location_archetype_id SET NOT NULL;
