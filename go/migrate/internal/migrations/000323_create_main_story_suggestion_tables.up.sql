CREATE TABLE IF NOT EXISTS main_story_templates (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    name text NOT NULL DEFAULT '',
    premise text NOT NULL DEFAULT '',
    district_fit text NOT NULL DEFAULT '',
    tone text NOT NULL DEFAULT '',
    theme_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    internal_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    faction_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    character_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    reveal_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    climax_summary text NOT NULL DEFAULT '',
    resolution_summary text NOT NULL DEFAULT '',
    why_it_works text NOT NULL DEFAULT '',
    beats jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS main_story_suggestion_jobs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    status text NOT NULL DEFAULT 'queued',
    count integer NOT NULL DEFAULT 1,
    quest_count integer NOT NULL DEFAULT 15,
    theme_prompt text NOT NULL DEFAULT '',
    district_fit text NOT NULL DEFAULT '',
    tone text NOT NULL DEFAULT '',
    family_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    character_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    internal_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    required_location_archetype_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    required_location_metadata_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_count integer NOT NULL DEFAULT 0,
    error_message text
);

CREATE TABLE IF NOT EXISTS main_story_suggestion_drafts (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    job_id uuid NOT NULL REFERENCES main_story_suggestion_jobs(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'suggested',
    name text NOT NULL DEFAULT '',
    premise text NOT NULL DEFAULT '',
    district_fit text NOT NULL DEFAULT '',
    tone text NOT NULL DEFAULT '',
    theme_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    internal_tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    faction_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    character_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    reveal_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    climax_summary text NOT NULL DEFAULT '',
    resolution_summary text NOT NULL DEFAULT '',
    why_it_works text NOT NULL DEFAULT '',
    beats jsonb NOT NULL DEFAULT '[]'::jsonb,
    warnings jsonb NOT NULL DEFAULT '[]'::jsonb,
    main_story_template_id uuid REFERENCES main_story_templates(id) ON DELETE SET NULL,
    converted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_main_story_suggestion_drafts_job_id
    ON main_story_suggestion_drafts(job_id);
