CREATE TABLE IF NOT EXISTS quest_archetype_suggestion_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL,
  count INTEGER NOT NULL DEFAULT 1,
  theme_prompt TEXT NOT NULL DEFAULT '',
  family_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  character_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  required_location_metadata_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS quest_archetype_suggestion_jobs_created_at_idx
  ON quest_archetype_suggestion_jobs(created_at DESC);

CREATE TABLE IF NOT EXISTS quest_archetype_suggestion_drafts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  job_id UUID NOT NULL REFERENCES quest_archetype_suggestion_jobs(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'suggested',
  name TEXT NOT NULL,
  hook TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  acceptance_dialogue JSONB NOT NULL DEFAULT '[]'::jsonb,
  character_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  difficulty_mode TEXT NOT NULL DEFAULT 'scale',
  difficulty INTEGER NOT NULL DEFAULT 1,
  monster_encounter_target_level INTEGER NOT NULL DEFAULT 1,
  why_this_scales TEXT NOT NULL DEFAULT '',
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  challenge_template_seeds JSONB NOT NULL DEFAULT '[]'::jsonb,
  scenario_template_seeds JSONB NOT NULL DEFAULT '[]'::jsonb,
  monster_template_seeds JSONB NOT NULL DEFAULT '[]'::jsonb,
  warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
  quest_archetype_id UUID REFERENCES quest_archetypes(id) ON DELETE SET NULL,
  converted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS quest_archetype_suggestion_drafts_job_id_idx
  ON quest_archetype_suggestion_drafts(job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS quest_archetype_suggestion_drafts_status_idx
  ON quest_archetype_suggestion_drafts(status);
