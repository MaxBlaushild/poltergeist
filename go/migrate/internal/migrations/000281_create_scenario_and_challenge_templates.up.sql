BEGIN;

CREATE TABLE IF NOT EXISTS scenario_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  prompt TEXT NOT NULL,
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE,
  reward_mode TEXT NOT NULL DEFAULT 'random',
  random_reward_size TEXT NOT NULL DEFAULT 'small',
  difficulty INTEGER NOT NULL DEFAULT 0,
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  open_ended BOOLEAN NOT NULL DEFAULT FALSE,
  failure_penalty_mode TEXT NOT NULL DEFAULT 'shared',
  failure_health_drain_type TEXT NOT NULL DEFAULT 'none',
  failure_health_drain_value INTEGER NOT NULL DEFAULT 0,
  failure_mana_drain_type TEXT NOT NULL DEFAULT 'none',
  failure_mana_drain_value INTEGER NOT NULL DEFAULT 0,
  failure_statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
  success_reward_mode TEXT NOT NULL DEFAULT 'shared',
  success_health_restore_type TEXT NOT NULL DEFAULT 'none',
  success_health_restore_value INTEGER NOT NULL DEFAULT 0,
  success_mana_restore_type TEXT NOT NULL DEFAULT 'none',
  success_mana_restore_value INTEGER NOT NULL DEFAULT 0,
  success_statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
  options JSONB NOT NULL DEFAULT '[]'::jsonb,
  item_rewards JSONB NOT NULL DEFAULT '[]'::jsonb,
  item_choice_rewards JSONB NOT NULL DEFAULT '[]'::jsonb,
  spell_rewards JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS challenge_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  location_archetype_id UUID NOT NULL REFERENCES location_archetypes(id) ON DELETE CASCADE,
  question TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  scale_with_user_level BOOLEAN NOT NULL DEFAULT FALSE,
  reward_mode TEXT NOT NULL DEFAULT 'random',
  random_reward_size TEXT NOT NULL DEFAULT 'small',
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward INTEGER NOT NULL DEFAULT 0,
  inventory_item_id INTEGER,
  item_choice_rewards JSONB NOT NULL DEFAULT '[]'::jsonb,
  submission_type TEXT NOT NULL DEFAULT 'photo',
  difficulty INTEGER NOT NULL DEFAULT 0,
  stat_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  proficiency TEXT
);

CREATE TABLE IF NOT EXISTS scenario_template_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL DEFAULT 'queued',
  count INTEGER NOT NULL DEFAULT 1,
  open_ended BOOLEAN NOT NULL DEFAULT FALSE,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE TABLE IF NOT EXISTS challenge_template_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  location_archetype_id UUID NOT NULL REFERENCES location_archetypes(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'queued',
  count INTEGER NOT NULL DEFAULT 1,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS challenge_templates_location_archetype_id_idx
  ON challenge_templates(location_archetype_id);

CREATE INDEX IF NOT EXISTS scenario_template_generation_jobs_created_at_idx
  ON scenario_template_generation_jobs(created_at DESC);

CREATE INDEX IF NOT EXISTS challenge_template_generation_jobs_location_archetype_id_idx
  ON challenge_template_generation_jobs(location_archetype_id);

CREATE INDEX IF NOT EXISTS challenge_template_generation_jobs_created_at_idx
  ON challenge_template_generation_jobs(created_at DESC);

COMMIT;
