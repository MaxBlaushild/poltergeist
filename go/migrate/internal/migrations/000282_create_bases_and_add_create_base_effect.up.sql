CREATE TABLE IF NOT EXISTS bases (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bases_user_id ON bases(user_id);
CREATE INDEX IF NOT EXISTS idx_bases_geometry ON bases USING GIST(geometry);

ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS consume_create_base BOOLEAN NOT NULL DEFAULT FALSE;
