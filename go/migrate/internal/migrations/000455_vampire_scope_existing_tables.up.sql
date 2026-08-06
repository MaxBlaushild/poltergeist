-- Vampire Ascendancy multi-tenancy, part 2: scope every play-state table to
-- an instance, and fold everything that exists in this database today into
-- one "legacy" instance so no data is lost. See 000454 for the
-- instances/admins tables and MULTI_TENANT_REQUIREMENTS.md for the plan.
--
-- created_by is left NULL on the legacy instance — there's no single
-- "creator" on record for an event that predates this concept. Grant a real
-- owner with `go run ./cmd/claim-owner` after this migration runs.
CREATE TEMP TABLE tmp_legacy_vampire_instance AS
  INSERT INTO vampire_instances (name, status, created_by)
  VALUES ('The Crimson Toast', 'active', NULL)
  RETURNING id;

-- vampire_game_state: was a singleton row (id always 1, CHECK id = 1); now
-- one row per instance, keyed by instance_id.
ALTER TABLE vampire_game_state ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_game_state SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_game_state ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_game_state DROP CONSTRAINT IF EXISTS vampire_game_state_singleton;
ALTER TABLE vampire_game_state DROP CONSTRAINT IF EXISTS vampire_game_state_pkey;
ALTER TABLE vampire_game_state DROP COLUMN IF EXISTS id;
ALTER TABLE vampire_game_state ADD PRIMARY KEY (instance_id);
ALTER TABLE vampire_game_state
  ADD CONSTRAINT vampire_game_state_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;

-- vampire_players
ALTER TABLE vampire_players ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_players SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_players ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_players
  ADD CONSTRAINT vampire_players_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_players_instance_idx ON vampire_players(instance_id);
-- token stays globally unique (opaque random, no collision risk) per
-- MULTI_TENANT_REQUIREMENTS.md's auth section.

-- vampire_games — name was globally UNIQUE; two instances must be able to
-- both run a game called e.g. "Bobbing for Blood", so that becomes
-- UNIQUE(instance_id, name) instead.
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_games SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_games ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_games
  ADD CONSTRAINT vampire_games_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
ALTER TABLE vampire_games DROP CONSTRAINT IF EXISTS vampire_games_name_key;
ALTER TABLE vampire_games ADD CONSTRAINT vampire_games_instance_name_key UNIQUE (instance_id, name);
CREATE INDEX IF NOT EXISTS vampire_games_instance_idx ON vampire_games(instance_id);

-- vampire_notifications
ALTER TABLE vampire_notifications ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_notifications SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_notifications ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_notifications
  ADD CONSTRAINT vampire_notifications_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_notifications_instance_idx ON vampire_notifications(instance_id);

-- vampire_house_favor_ledger
ALTER TABLE vampire_house_favor_ledger ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_house_favor_ledger SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_house_favor_ledger ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_house_favor_ledger
  ADD CONSTRAINT vampire_house_favor_ledger_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_house_favor_ledger_instance_idx ON vampire_house_favor_ledger(instance_id);

-- vampire_blood_token_log
ALTER TABLE vampire_blood_token_log ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_blood_token_log SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_blood_token_log ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_blood_token_log
  ADD CONSTRAINT vampire_blood_token_log_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_blood_token_log_instance_idx ON vampire_blood_token_log(instance_id);

-- The archive tables were created as `LIKE vampire_..._ledger` (a one-time
-- snapshot of columns, not a tracking view), so they don't pick up the
-- instance_id column just added to their live counterparts. Add it
-- explicitly, or ResetGameProgress's archive-then-wipe breaks the next time
-- it runs. (vampire.go's archive step uses an explicit column list, not
-- `SELECT *`, so column order between live and archive doesn't need to
-- match.)
ALTER TABLE vampire_house_favor_ledger_archive ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_house_favor_ledger_archive SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_house_favor_ledger_archive ALTER COLUMN instance_id SET NOT NULL;

ALTER TABLE vampire_blood_token_log_archive ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_blood_token_log_archive SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_blood_token_log_archive ALTER COLUMN instance_id SET NOT NULL;

-- vampire_mission_submissions — technically derivable via player_id, but
-- kept direct for simple scoped queries (matches the ledgers above).
ALTER TABLE vampire_mission_submissions ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_mission_submissions SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_mission_submissions ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_mission_submissions
  ADD CONSTRAINT vampire_mission_submissions_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_mission_submissions_instance_idx ON vampire_mission_submissions(instance_id);

-- vampire_quiz_submissions
ALTER TABLE vampire_quiz_submissions ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_quiz_submissions SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_quiz_submissions ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_quiz_submissions
  ADD CONSTRAINT vampire_quiz_submissions_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_quiz_submissions_instance_idx ON vampire_quiz_submissions(instance_id);

-- vampire_gm_action_log
ALTER TABLE vampire_gm_action_log ADD COLUMN IF NOT EXISTS instance_id UUID;
UPDATE vampire_gm_action_log SET instance_id = (SELECT id FROM tmp_legacy_vampire_instance) WHERE instance_id IS NULL;
ALTER TABLE vampire_gm_action_log ALTER COLUMN instance_id SET NOT NULL;
ALTER TABLE vampire_gm_action_log
  ADD CONSTRAINT vampire_gm_action_log_instance_fk
  FOREIGN KEY (instance_id) REFERENCES vampire_instances(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS vampire_gm_action_log_instance_idx ON vampire_gm_action_log(instance_id);
