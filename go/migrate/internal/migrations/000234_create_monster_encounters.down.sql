ALTER TABLE quest_nodes
  DROP CONSTRAINT IF EXISTS quest_nodes_exactly_one_target;

UPDATE quest_nodes qn
SET monster_id = first_member.monster_id
FROM LATERAL (
  SELECT mem.monster_id
  FROM monster_encounter_members mem
  WHERE mem.monster_encounter_id = qn.monster_encounter_id
  ORDER BY mem.slot ASC, mem.created_at ASC
  LIMIT 1
) AS first_member
WHERE qn.monster_id IS NULL
  AND qn.monster_encounter_id IS NOT NULL;

DROP INDEX IF EXISTS quest_nodes_monster_encounter_idx;

ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS monster_encounter_id;

DROP TABLE IF EXISTS monster_encounter_members;
DROP TABLE IF EXISTS monster_encounters;

ALTER TABLE quest_nodes
  ADD CONSTRAINT quest_nodes_exactly_one_target
  CHECK (
    (
      (CASE WHEN scenario_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN monster_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN challenge_id IS NOT NULL THEN 1 ELSE 0 END)
    ) = 1
  );
