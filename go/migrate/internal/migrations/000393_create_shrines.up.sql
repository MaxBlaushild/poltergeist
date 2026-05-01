BEGIN;

ALTER TABLE zone_kinds
  ADD COLUMN IF NOT EXISTS shrine_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1;

ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS shrine_count INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS shrine_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_kind TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  blessing_name TEXT NOT NULL DEFAULT '',
  effect_description TEXT NOT NULL DEFAULT '',
  effect_kind TEXT NOT NULL DEFAULT 'strength',
  base_magnitude INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS shrine_templates_zone_kind_idx
  ON shrine_templates(zone_kind);

CREATE TABLE IF NOT EXISTS shrine_template_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_kind TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'queued',
  count INTEGER NOT NULL DEFAULT 1,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS shrine_template_generation_jobs_created_at_idx
  ON shrine_template_generation_jobs(created_at DESC);

CREATE INDEX IF NOT EXISTS shrine_template_generation_jobs_zone_kind_idx
  ON shrine_template_generation_jobs(zone_kind);

CREATE TABLE IF NOT EXISTS shrines (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  shrine_template_id UUID NOT NULL REFERENCES shrine_templates(id) ON DELETE CASCADE,
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  zone_kind TEXT NOT NULL DEFAULT '',
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326),
  cooldown_seconds INTEGER NOT NULL DEFAULT 604800,
  invalidated BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS shrines_zone_id_idx
  ON shrines(zone_id);

CREATE INDEX IF NOT EXISTS shrines_template_id_idx
  ON shrines(shrine_template_id);

CREATE INDEX IF NOT EXISTS shrines_geometry_idx
  ON shrines USING GIST(geometry);

CREATE TABLE IF NOT EXISTS user_shrine_uses (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  shrine_id UUID NOT NULL REFERENCES shrines(id) ON DELETE CASCADE,
  used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS user_shrine_uses_user_id_idx
  ON user_shrine_uses(user_id);

CREATE INDEX IF NOT EXISTS user_shrine_uses_shrine_id_idx
  ON user_shrine_uses(shrine_id);

CREATE INDEX IF NOT EXISTS user_shrine_uses_user_used_at_idx
  ON user_shrine_uses(user_id, used_at DESC);

CREATE INDEX IF NOT EXISTS user_shrine_uses_user_shrine_used_at_idx
  ON user_shrine_uses(user_id, shrine_id, used_at DESC);

COMMIT;
