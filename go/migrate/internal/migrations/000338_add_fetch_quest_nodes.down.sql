ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS fetch_requirements_json,
  DROP COLUMN IF EXISTS fetch_character_id;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS fetch_requirements_json,
  DROP COLUMN IF EXISTS fetch_character_id;
