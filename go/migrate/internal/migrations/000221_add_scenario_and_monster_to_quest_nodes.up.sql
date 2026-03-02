ALTER TABLE quest_nodes
  ADD COLUMN scenario_id UUID REFERENCES scenarios(id) ON DELETE SET NULL,
  ADD COLUMN monster_id UUID REFERENCES monsters(id) ON DELETE SET NULL;

CREATE INDEX quest_nodes_scenario_idx ON quest_nodes(scenario_id);
CREATE INDEX quest_nodes_monster_idx ON quest_nodes(monster_id);
