UPDATE quest_archetype_nodes
SET node_type = CASE
  WHEN node_type = 'challenge' THEN 'location'
  ELSE node_type
END
WHERE TRUE;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS location_selection_mode,
  DROP COLUMN IF EXISTS challenge_template_id;
