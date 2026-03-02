DROP INDEX IF EXISTS quest_nodes_monster_idx;
DROP INDEX IF EXISTS quest_nodes_scenario_idx;

ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS monster_id,
  DROP COLUMN IF EXISTS scenario_id;
