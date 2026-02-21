DROP TABLE IF EXISTS user_proficiencies;

ALTER TABLE quest_node_challenges
DROP COLUMN IF EXISTS proficiency;
