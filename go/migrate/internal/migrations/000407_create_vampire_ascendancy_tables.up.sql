-- Vampire Ascendancy (The Crimson Toast) event app schema.
-- All tables are prefixed vampire_ to keep this self-contained.

CREATE TABLE IF NOT EXISTS vampire_houses (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL UNIQUE,
  sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS vampire_characters (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL DEFAULT '',
  house_id UUID REFERENCES vampire_houses(id) ON DELETE SET NULL,
  role_type TEXT NOT NULL DEFAULT 'player', -- player | gm | npc
  is_optional BOOLEAN NOT NULL DEFAULT FALSE,
  pre_event_info TEXT NOT NULL DEFAULT '',
  post_act1_context TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS vampire_secrets (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  ordinal INTEGER NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  UNIQUE (character_id, ordinal)
);

CREATE TABLE IF NOT EXISTS vampire_missions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  ordinal INTEGER NOT NULL,
  tier TEXT NOT NULL DEFAULT 'easy', -- easy | medium | hard
  reward_bt INTEGER NOT NULL DEFAULT 0,
  prompt TEXT NOT NULL DEFAULT '',
  answer_format TEXT NOT NULL DEFAULT '',
  UNIQUE (character_id, ordinal)
);

-- A player is a physical guest, authenticated by an opaque per-character token.
-- character_id is editable by GMs up to event start.
CREATE TABLE IF NOT EXISTS vampire_players (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  token TEXT NOT NULL UNIQUE,
  character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL,
  guest_label TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS vampire_players_token_idx ON vampire_players(token);
CREATE INDEX IF NOT EXISTS vampire_players_character_idx ON vampire_players(character_id);

-- Absence of a row means not_started; rows are created when a player submits.
CREATE TABLE IF NOT EXISTS vampire_mission_submissions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  player_id UUID NOT NULL REFERENCES vampire_players(id) ON DELETE CASCADE,
  mission_id UUID NOT NULL REFERENCES vampire_missions(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'submitted', -- submitted | verified | rejected
  player_answer TEXT NOT NULL DEFAULT '',
  awarded_bt INTEGER NOT NULL DEFAULT 0,
  verified_by TEXT NOT NULL DEFAULT '',
  UNIQUE (player_id, mission_id)
);

-- Append-only ledger. Leaderboard = SUM(delta) grouped by house.
CREATE TABLE IF NOT EXISTS vampire_house_favor_ledger (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  house_id UUID NOT NULL REFERENCES vampire_houses(id) ON DELETE CASCADE,
  delta INTEGER NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  gm_name TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'manual' -- manual | mission | quiz
);
CREATE INDEX IF NOT EXISTS vampire_house_favor_ledger_house_idx ON vampire_house_favor_ledger(house_id);

-- Blood Tokens are physical; this log records awards for reference/backup tally.
CREATE TABLE IF NOT EXISTS vampire_blood_token_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  player_id UUID NOT NULL REFERENCES vampire_players(id) ON DELETE CASCADE,
  delta INTEGER NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'manual', -- manual | mission | physical_game
  gm_name TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS vampire_blood_token_log_player_idx ON vampire_blood_token_log(player_id);

-- Singleton global game state (id is always 1).
CREATE TABLE IF NOT EXISTS vampire_game_state (
  id INTEGER PRIMARY KEY DEFAULT 1,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  current_act TEXT NOT NULL DEFAULT 'pre_event', -- pre_event | act1 | act2 | act3 | quiz | resolved
  content_unlocked BOOLEAN NOT NULL DEFAULT FALSE,
  quiz_open BOOLEAN NOT NULL DEFAULT FALSE,
  active_notification_id UUID,
  CONSTRAINT vampire_game_state_singleton CHECK (id = 1)
);
INSERT INTO vampire_game_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS vampire_notifications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  title TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT 'all', -- all | house | player
  target_id UUID, -- house_id or player_id depending on scope
  created_by TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Data-driven quiz; hf_effect maps house name -> delta applied on a correct answer.
CREATE TABLE IF NOT EXISTS vampire_quiz_questions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  ordinal INTEGER NOT NULL DEFAULT 0,
  prompt TEXT NOT NULL DEFAULT '',
  question_type TEXT NOT NULL DEFAULT 'open', -- multiple_choice | open
  options JSONB NOT NULL DEFAULT '[]'::jsonb,
  correct_answer TEXT NOT NULL DEFAULT '',
  hf_effect JSONB NOT NULL DEFAULT '{}'::jsonb,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS vampire_quiz_submissions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  player_id UUID NOT NULL REFERENCES vampire_players(id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES vampire_quiz_questions(id) ON DELETE CASCADE,
  answer TEXT NOT NULL DEFAULT '',
  is_correct BOOLEAN,
  locked BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE (player_id, question_id)
);

-- Audit trail so the four GMs do not step on each other.
CREATE TABLE IF NOT EXISTS vampire_gm_action_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  gm_name TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb
);
