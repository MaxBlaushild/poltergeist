ALTER TABLE quest_node_children
  DROP COLUMN IF EXISTS quest_node_challenge_id;

DROP TABLE IF EXISTS quest_node_challenges CASCADE;

DROP TABLE IF EXISTS quest_node_child_legacy_challenge_link_archive;
DROP TABLE IF EXISTS quest_node_challenge_legacy_archive;

DROP FUNCTION IF EXISTS prevent_quest_node_challenges_on_challenge_nodes();
DROP FUNCTION IF EXISTS prevent_challenge_targets_on_nodes_with_prompts();
