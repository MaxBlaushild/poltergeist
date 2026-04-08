DROP INDEX IF EXISTS quest_nodes_exposition_idx;
ALTER TABLE quest_nodes DROP COLUMN IF EXISTS exposition_id;

DROP INDEX IF EXISTS idx_user_exposition_completions_exposition_id;
DROP INDEX IF EXISTS idx_user_exposition_completions_user_id;
DROP TABLE IF EXISTS user_exposition_completions;

DROP INDEX IF EXISTS idx_exposition_spell_rewards_spell_id;
DROP INDEX IF EXISTS idx_exposition_spell_rewards_exposition_id;
DROP TABLE IF EXISTS exposition_spell_rewards;

DROP INDEX IF EXISTS idx_exposition_item_rewards_exposition_id;
DROP TABLE IF EXISTS exposition_item_rewards;

DROP INDEX IF EXISTS idx_expositions_point_of_interest_id;
DROP INDEX IF EXISTS idx_expositions_geometry;
DROP INDEX IF EXISTS idx_expositions_zone_id;
DROP TABLE IF EXISTS expositions;
