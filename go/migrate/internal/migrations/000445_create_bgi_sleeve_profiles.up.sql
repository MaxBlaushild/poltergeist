-- R-2.2: named sleeve classes, not a free-text thickness number. Total
-- per-card thickness (base + 2x sleeve, since a sleeve wraps both faces) is
-- computed in Go (BgiSleeveProfile.TotalCardThicknessMm) rather than stored,
-- so that arithmetic lives in exactly one tested place.
CREATE TABLE IF NOT EXISTS bgi_sleeve_profiles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  class_key TEXT NOT NULL UNIQUE,
  label TEXT NOT NULL,
  base_card_thickness_mm NUMERIC NOT NULL,
  sleeve_material_thickness_mm NUMERIC NOT NULL DEFAULT 0,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  source_url TEXT NOT NULL DEFAULT '',
  CONSTRAINT bgi_sleeve_profiles_verified_requires_source CHECK (NOT verified OR source_url <> '')
);
