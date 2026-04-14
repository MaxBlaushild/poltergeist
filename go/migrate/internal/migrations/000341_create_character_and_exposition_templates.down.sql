DROP INDEX IF EXISTS idx_tutorial_configs_base_quest_giver_character_template_id;
DROP INDEX IF EXISTS idx_quest_archetype_nodes_exposition_template_id;
DROP INDEX IF EXISTS idx_quest_archetype_nodes_fetch_character_template_id;

ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS base_quest_giver_character_template_id;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS exposition_template_id,
  DROP COLUMN IF EXISTS fetch_character_template_id;

DROP TABLE IF EXISTS exposition_templates;
DROP TABLE IF EXISTS character_templates;
