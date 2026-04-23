CREATE TABLE IF NOT EXISTS monster_template_suggestion_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL DEFAULT 'queued',
  monster_type TEXT NOT NULL DEFAULT 'monster',
  genre_id UUID REFERENCES zone_genres(id) ON DELETE SET NULL,
  zone_kind TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'seed_generated',
  count INTEGER NOT NULL DEFAULT 1,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_monster_template_suggestion_jobs_created_at
  ON monster_template_suggestion_jobs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_monster_template_suggestion_jobs_status_created_at
  ON monster_template_suggestion_jobs (status, created_at DESC);

CREATE TABLE IF NOT EXISTS monster_template_suggestion_drafts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  job_id UUID NOT NULL REFERENCES monster_template_suggestion_jobs(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'suggested',
  monster_type TEXT NOT NULL DEFAULT 'monster',
  genre_id UUID REFERENCES zone_genres(id) ON DELETE SET NULL,
  zone_kind TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  monster_template_id UUID REFERENCES monster_templates(id) ON DELETE SET NULL,
  converted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_monster_template_suggestion_drafts_job_id_created_at
  ON monster_template_suggestion_drafts (job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_monster_template_suggestion_drafts_status
  ON monster_template_suggestion_drafts (status);

CREATE INDEX IF NOT EXISTS idx_monster_template_suggestion_drafts_monster_template_id
  ON monster_template_suggestion_drafts (monster_template_id);
