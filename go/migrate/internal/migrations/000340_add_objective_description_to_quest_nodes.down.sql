ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS objective_description;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS objective_description;
