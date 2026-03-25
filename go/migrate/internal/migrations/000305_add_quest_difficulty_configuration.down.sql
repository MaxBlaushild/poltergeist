ALTER TABLE quest_node_challenges
  DROP COLUMN IF EXISTS scale_with_user_level;

ALTER TABLE quests
  DROP COLUMN IF EXISTS difficulty_mode,
  DROP COLUMN IF EXISTS difficulty;

ALTER TABLE quest_archetypes
  DROP COLUMN IF EXISTS difficulty_mode,
  DROP COLUMN IF EXISTS difficulty;
