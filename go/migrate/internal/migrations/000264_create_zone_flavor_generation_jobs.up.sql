BEGIN;

CREATE TABLE IF NOT EXISTS zone_flavor_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'queued',
  generated_description TEXT,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS zone_flavor_generation_jobs_zone_id_idx
  ON zone_flavor_generation_jobs(zone_id);

CREATE INDEX IF NOT EXISTS zone_flavor_generation_jobs_status_idx
  ON zone_flavor_generation_jobs(status);

CREATE INDEX IF NOT EXISTS zone_flavor_generation_jobs_created_at_idx
  ON zone_flavor_generation_jobs(created_at DESC);

COMMIT;
