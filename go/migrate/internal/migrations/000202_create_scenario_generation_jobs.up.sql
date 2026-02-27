BEGIN;

CREATE TABLE scenario_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'queued',
  open_ended BOOLEAN NOT NULL DEFAULT FALSE,
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  generated_scenario_id UUID REFERENCES scenarios(id) ON DELETE SET NULL,
  error_message TEXT
);

CREATE INDEX scenario_generation_jobs_zone_id_idx
  ON scenario_generation_jobs(zone_id);

CREATE INDEX scenario_generation_jobs_status_idx
  ON scenario_generation_jobs(status);

CREATE INDEX scenario_generation_jobs_created_at_idx
  ON scenario_generation_jobs(created_at DESC);

COMMIT;
