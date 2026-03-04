BEGIN;

CREATE TABLE IF NOT EXISTS challenge_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'queued',
  count INTEGER NOT NULL DEFAULT 1,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS challenge_generation_jobs_zone_id_idx
  ON challenge_generation_jobs(zone_id);

CREATE INDEX IF NOT EXISTS challenge_generation_jobs_status_idx
  ON challenge_generation_jobs(status);

CREATE INDEX IF NOT EXISTS challenge_generation_jobs_created_at_idx
  ON challenge_generation_jobs(created_at DESC);

COMMIT;
