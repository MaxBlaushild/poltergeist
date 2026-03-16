ALTER TABLE monster_battles
  ADD COLUMN IF NOT EXISTS monster_encounter_id UUID REFERENCES monster_encounters(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS monster_battles_monster_encounter_id_idx
  ON monster_battles(monster_encounter_id);

UPDATE monster_battles mb
SET monster_encounter_id = sub.monster_encounter_id
FROM (
  SELECT DISTINCT ON (mem.monster_id)
    mem.monster_id,
    mem.monster_encounter_id
  FROM monster_encounter_members mem
  JOIN monster_encounters me ON me.id = mem.monster_encounter_id
  WHERE me.retired_at IS NULL
  ORDER BY mem.monster_id, me.created_at ASC
) AS sub
WHERE mb.monster_encounter_id IS NULL
  AND mb.monster_id = sub.monster_id;
