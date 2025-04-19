DROP INDEX IF EXISTS idx_quest_archtype_type_node_id;
DROP INDEX IF EXISTS idx_quest_archtype_type_challenge_id;
DROP INDEX IF EXISTS idx_quest_archtype_challenge_node_id;
DROP INDEX IF EXISTS idx_quest_archtype_challenge_unlocked_node_id;

DROP TABLE IF EXISTS QuestArchtypes;
DROP TABLE IF EXISTS QuestArchTypeNodeChallenges;
DROP TABLE IF EXISTS QuestArchtypeChallenges;
DROP TABLE IF EXISTS QuestArchtypeNodes;
