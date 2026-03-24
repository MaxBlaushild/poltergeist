ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS monster_template_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE quest_archetype_nodes AS q
SET monster_template_ids = resolved.template_ids
FROM (
  SELECT
    qn.id,
    COALESCE(
      jsonb_agg(DISTINCT m.template_id::text) FILTER (WHERE m.template_id IS NOT NULL),
      '[]'::jsonb
    ) AS template_ids
  FROM quest_archetype_nodes AS qn
  LEFT JOIN LATERAL jsonb_array_elements_text(COALESCE(qn.monster_ids, '[]'::jsonb)) AS source(monster_id)
    ON TRUE
  LEFT JOIN monsters AS m
    ON m.id::text = source.monster_id
  GROUP BY qn.id
) AS resolved
WHERE q.id = resolved.id;
