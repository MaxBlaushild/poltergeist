CREATE TABLE zone_kinds (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  slug TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  place_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  monster_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  boss_encounter_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  raid_encounter_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  input_encounter_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  option_encounter_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  treasure_chest_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  healing_fountain_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1,
  resource_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX idx_zone_kinds_slug_unique ON zone_kinds (slug);
CREATE INDEX idx_zone_kinds_name_slug ON zone_kinds (name ASC, slug ASC);
