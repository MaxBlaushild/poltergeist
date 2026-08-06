-- Reverses 000455. Lossy if more than one instance/game_state row exists by
-- the time this runs (multi-tenancy is being un-done, so only one instance's
-- worth of state can survive) — not expected to be run against real
-- multi-instance data, only as a development rollback.

ALTER TABLE vampire_gm_action_log DROP CONSTRAINT IF EXISTS vampire_gm_action_log_instance_fk;
DROP INDEX IF EXISTS vampire_gm_action_log_instance_idx;
ALTER TABLE vampire_gm_action_log DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_quiz_submissions DROP CONSTRAINT IF EXISTS vampire_quiz_submissions_instance_fk;
DROP INDEX IF EXISTS vampire_quiz_submissions_instance_idx;
ALTER TABLE vampire_quiz_submissions DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_mission_submissions DROP CONSTRAINT IF EXISTS vampire_mission_submissions_instance_fk;
DROP INDEX IF EXISTS vampire_mission_submissions_instance_idx;
ALTER TABLE vampire_mission_submissions DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_blood_token_log DROP CONSTRAINT IF EXISTS vampire_blood_token_log_instance_fk;
DROP INDEX IF EXISTS vampire_blood_token_log_instance_idx;
ALTER TABLE vampire_blood_token_log DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_house_favor_ledger_archive DROP COLUMN IF EXISTS instance_id;
ALTER TABLE vampire_blood_token_log_archive DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_house_favor_ledger DROP CONSTRAINT IF EXISTS vampire_house_favor_ledger_instance_fk;
DROP INDEX IF EXISTS vampire_house_favor_ledger_instance_idx;
ALTER TABLE vampire_house_favor_ledger DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_notifications DROP CONSTRAINT IF EXISTS vampire_notifications_instance_fk;
DROP INDEX IF EXISTS vampire_notifications_instance_idx;
ALTER TABLE vampire_notifications DROP COLUMN IF EXISTS instance_id;

DROP INDEX IF EXISTS vampire_games_instance_idx;
ALTER TABLE vampire_games DROP CONSTRAINT IF EXISTS vampire_games_instance_name_key;
ALTER TABLE vampire_games ADD CONSTRAINT vampire_games_name_key UNIQUE (name);
ALTER TABLE vampire_games DROP CONSTRAINT IF EXISTS vampire_games_instance_fk;
ALTER TABLE vampire_games DROP COLUMN IF EXISTS instance_id;

ALTER TABLE vampire_players DROP CONSTRAINT IF EXISTS vampire_players_instance_fk;
DROP INDEX IF EXISTS vampire_players_instance_idx;
ALTER TABLE vampire_players DROP COLUMN IF EXISTS instance_id;

-- vampire_game_state: collapse back to a singleton keyed by id = 1. If
-- multiple instances exist, only one row survives (arbitrary pick) — see
-- note above.
DELETE FROM vampire_game_state
  WHERE instance_id NOT IN (SELECT instance_id FROM vampire_game_state ORDER BY updated_at DESC LIMIT 1);
ALTER TABLE vampire_game_state DROP CONSTRAINT IF EXISTS vampire_game_state_instance_fk;
ALTER TABLE vampire_game_state DROP CONSTRAINT IF EXISTS vampire_game_state_pkey;
ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS id INTEGER;
UPDATE vampire_game_state SET id = 1;
ALTER TABLE vampire_game_state ALTER COLUMN id SET DEFAULT 1;
ALTER TABLE vampire_game_state ALTER COLUMN id SET NOT NULL;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS instance_id;
ALTER TABLE vampire_game_state ADD PRIMARY KEY (id);
ALTER TABLE vampire_game_state ADD CONSTRAINT vampire_game_state_singleton CHECK (id = 1);

DELETE FROM vampire_instances WHERE name = 'The Crimson Toast' AND created_by IS NULL;
