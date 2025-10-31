BEGIN;

DROP INDEX IF EXISTS idx_character_actions_action_type;
DROP INDEX IF EXISTS idx_character_actions_character_id;
DROP TABLE IF EXISTS character_actions;

COMMIT;

