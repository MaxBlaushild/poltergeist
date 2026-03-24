CREATE TABLE district_seed_jobs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    district_id uuid NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    status text NOT NULL,
    error_message text,
    quest_archetype_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    results jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX idx_district_seed_jobs_district_id_created_at
    ON district_seed_jobs (district_id, created_at DESC);

CREATE INDEX idx_district_seed_jobs_status
    ON district_seed_jobs (status);
