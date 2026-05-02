CREATE TABLE reward_profiles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  slug TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  preferred_item_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  preferred_material_keys JSONB NOT NULL DEFAULT '[]'::jsonb,
  preferred_damage_affinities JSONB NOT NULL DEFAULT '[]'::jsonb,
  preferred_resource_type_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  prefer_equipment BOOLEAN NOT NULL DEFAULT FALSE,
  prefer_utility BOOLEAN NOT NULL DEFAULT FALSE,
  prefer_knowledge BOOLEAN NOT NULL DEFAULT FALSE,
  prefer_non_equipment BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_reward_profiles_slug_unique ON reward_profiles (slug);
CREATE INDEX idx_reward_profiles_name_slug ON reward_profiles (name ASC, slug ASC);
