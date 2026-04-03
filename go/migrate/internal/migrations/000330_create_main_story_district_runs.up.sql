CREATE TABLE IF NOT EXISTS main_story_district_runs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    main_story_template_id uuid NOT NULL REFERENCES main_story_templates(id) ON DELETE CASCADE,
    district_id uuid NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'in_progress',
    beat_runs jsonb NOT NULL DEFAULT '[]'::jsonb,
    generated_character_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    error_message text
);

CREATE INDEX IF NOT EXISTS idx_main_story_district_runs_template_id
    ON main_story_district_runs(main_story_template_id);

CREATE INDEX IF NOT EXISTS idx_main_story_district_runs_district_id
    ON main_story_district_runs(district_id);
