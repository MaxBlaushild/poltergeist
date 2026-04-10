ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS objective_description TEXT NOT NULL DEFAULT '';

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS objective_description TEXT NOT NULL DEFAULT '';

UPDATE quest_archetype_nodes
SET objective_description = BTRIM(objective_description)
WHERE objective_description IS NOT NULL;

UPDATE quest_nodes
SET objective_description = BTRIM(objective_description)
WHERE objective_description IS NOT NULL;
