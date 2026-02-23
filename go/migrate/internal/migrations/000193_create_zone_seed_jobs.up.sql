BEGIN;

CREATE TABLE zone_seed_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    zone_id UUID NOT NULL REFERENCES zones(id),
    status TEXT NOT NULL,
    error_message TEXT,
    place_count INT NOT NULL DEFAULT 0,
    character_count INT NOT NULL DEFAULT 0,
    quest_count INT NOT NULL DEFAULT 0,
    draft JSONB
);

CREATE INDEX idx_zone_seed_jobs_zone_id ON zone_seed_jobs(zone_id);
CREATE INDEX idx_zone_seed_jobs_status ON zone_seed_jobs(status);

COMMIT;
