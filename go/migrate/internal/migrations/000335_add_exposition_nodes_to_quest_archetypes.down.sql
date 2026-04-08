DELETE FROM quest_archetype_node_challenges
WHERE quest_archetype_node_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'exposition'
)
OR quest_archetype_challenge_id IN (
  SELECT id
  FROM quest_archetype_challenges
  WHERE unlocked_node_id IN (
    SELECT id FROM quest_archetype_nodes WHERE node_type = 'exposition'
  )
);

DELETE FROM quest_archetype_challenges
WHERE unlocked_node_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'exposition'
);

DELETE FROM quest_archetypes
WHERE root_id IN (
  SELECT id FROM quest_archetype_nodes WHERE node_type = 'exposition'
);

DELETE FROM quest_archetype_nodes
WHERE node_type = 'exposition';

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS exposition_spell_rewards_json,
  DROP COLUMN IF EXISTS exposition_item_rewards_json,
  DROP COLUMN IF EXISTS exposition_material_rewards_json,
  DROP COLUMN IF EXISTS exposition_reward_gold,
  DROP COLUMN IF EXISTS exposition_reward_experience,
  DROP COLUMN IF EXISTS exposition_random_reward_size,
  DROP COLUMN IF EXISTS exposition_reward_mode,
  DROP COLUMN IF EXISTS exposition_dialogue,
  DROP COLUMN IF EXISTS exposition_description,
  DROP COLUMN IF EXISTS exposition_title;
