ALTER TABLE bases
ADD COLUMN IF NOT EXISTS description TEXT;

CREATE TABLE IF NOT EXISTS base_description_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  generated_description TEXT,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS base_description_generation_jobs_base_id_idx
  ON base_description_generation_jobs(base_id);

CREATE INDEX IF NOT EXISTS base_description_generation_jobs_status_idx
  ON base_description_generation_jobs(status);

CREATE INDEX IF NOT EXISTS base_description_generation_jobs_created_at_idx
  ON base_description_generation_jobs(created_at DESC);
