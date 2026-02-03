DROP INDEX IF EXISTS quest_node_progress_node_idx;
DROP INDEX IF EXISTS quest_node_progress_acceptance_idx;
DROP TABLE quest_node_progress;

DROP INDEX IF EXISTS quest_acceptances_v2_quest_idx;
DROP INDEX IF EXISTS quest_acceptances_v2_user_idx;
DROP TABLE quest_acceptances_v2;

DROP INDEX IF EXISTS quest_node_challenges_node_idx;
DROP TABLE quest_node_challenges;

DROP INDEX IF EXISTS quest_node_children_next_idx;
DROP INDEX IF EXISTS quest_node_children_node_idx;
DROP TABLE quest_node_children;

DROP INDEX IF EXISTS quest_nodes_poi_idx;
DROP INDEX IF EXISTS quest_nodes_quest_id_idx;
DROP TABLE quest_nodes;

DROP TABLE quests;
