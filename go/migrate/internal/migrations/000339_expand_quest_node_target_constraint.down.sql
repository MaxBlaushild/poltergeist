ALTER TABLE quest_nodes
  DROP CONSTRAINT IF EXISTS quest_nodes_exactly_one_target;

ALTER TABLE quest_nodes
  ADD CONSTRAINT quest_nodes_exactly_one_target
  CHECK (
    (
      (CASE WHEN scenario_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN challenge_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN monster_id IS NOT NULL OR monster_encounter_id IS NOT NULL THEN 1 ELSE 0 END)
    ) = 1
  );
