ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS story_flag_key;

ALTER TABLE quest_archetype_nodes
  DROP COLUMN IF EXISTS story_flag_key;
