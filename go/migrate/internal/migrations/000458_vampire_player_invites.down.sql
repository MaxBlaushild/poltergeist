COMMENT ON COLUMN vampire_instance_characters.included IS NULL;
COMMENT ON COLUMN vampire_instance_characters.sigil IS NULL;

COMMENT ON COLUMN vampire_players.token IS NULL;
ALTER TABLE vampire_players ALTER COLUMN token SET NOT NULL;

DROP INDEX IF EXISTS vampire_players_instance_user_idx;
ALTER TABLE vampire_players DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS vampire_player_invites;
