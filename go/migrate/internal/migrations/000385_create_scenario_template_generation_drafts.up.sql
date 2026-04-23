CREATE TABLE IF NOT EXISTS scenario_template_generation_drafts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  job_id UUID NOT NULL REFERENCES scenario_template_generation_jobs(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'suggested',
  genre_id UUID NOT NULL REFERENCES zone_genres(id),
  zone_kind TEXT NOT NULL DEFAULT '',
  prompt TEXT NOT NULL DEFAULT '',
  open_ended BOOLEAN NOT NULL DEFAULT FALSE,
  difficulty INTEGER NOT NULL DEFAULT 0,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  scenario_template_id UUID REFERENCES scenario_templates(id) ON DELETE SET NULL,
  converted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS scenario_template_generation_drafts_job_id_idx
  ON scenario_template_generation_drafts(job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS scenario_template_generation_drafts_status_idx
  ON scenario_template_generation_drafts(status);
