DROP INDEX IF EXISTS idx_quest_archtype_type_node_id;
DROP INDEX IF EXISTS idx_quest_archtype_type_challenge_id;
DROP INDEX IF EXISTS idx_quest_archtype_challenge_node_id;
DROP INDEX IF EXISTS idx_quest_archtype_challenge_unlocked_node_id;

DROP TABLE IF EXISTS quest_archetypes;
DROP TABLE IF EXISTS quest_archetype_node_challenges;
DROP TABLE IF EXISTS quest_archetype_challenges;
DROP TABLE IF EXISTS quest_archetype_nodes;
