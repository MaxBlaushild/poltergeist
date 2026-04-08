ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS story_flag_key TEXT NOT NULL DEFAULT '';

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS story_flag_key TEXT NOT NULL DEFAULT '';

UPDATE quest_archetype_nodes
SET story_flag_key = LOWER(BTRIM(story_flag_key))
WHERE story_flag_key IS NOT NULL;

UPDATE quest_nodes
SET story_flag_key = LOWER(BTRIM(story_flag_key))
WHERE story_flag_key IS NOT NULL;
