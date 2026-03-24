CREATE TABLE districts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT ''
);

CREATE INDEX districts_deleted_at_idx ON districts (deleted_at);

CREATE TABLE district_zones (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  CONSTRAINT district_zones_district_zone_unique UNIQUE (district_id, zone_id)
);

CREATE INDEX district_zones_district_id_idx ON district_zones (district_id);
CREATE INDEX district_zones_zone_id_idx ON district_zones (zone_id);
