-- Lossy: any character a player chose for themselves after this migration
-- has no invite row to attribute a character_id back to.
ALTER TABLE vampire_player_invites ADD COLUMN IF NOT EXISTS character_id UUID REFERENCES vampire_characters(id);
DROP TABLE IF EXISTS vampire_instance_character_pool;
