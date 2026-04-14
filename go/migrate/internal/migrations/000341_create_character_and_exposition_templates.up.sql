CREATE TABLE IF NOT EXISTS character_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  map_icon_url TEXT NOT NULL DEFAULT '',
  dialogue_image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  story_variants JSONB NOT NULL DEFAULT '[]'::jsonb,
  image_generation_status TEXT NOT NULL DEFAULT 'none',
  image_generation_error TEXT
);

CREATE TABLE IF NOT EXISTS exposition_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  title TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  dialogue JSONB NOT NULL DEFAULT '[]'::jsonb,
  required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb,
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  reward_mode TEXT NOT NULL DEFAULT 'random',
  random_reward_size TEXT NOT NULL DEFAULT 'small',
  reward_experience INT NOT NULL DEFAULT 0,
  reward_gold INT NOT NULL DEFAULT 0,
  material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  item_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  spell_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb
);

ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS fetch_character_template_id UUID REFERENCES character_templates(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS exposition_template_id UUID REFERENCES exposition_templates(id) ON DELETE SET NULL;

ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS base_quest_giver_character_template_id UUID REFERENCES character_templates(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_quest_archetype_nodes_fetch_character_template_id
  ON quest_archetype_nodes(fetch_character_template_id);

CREATE INDEX IF NOT EXISTS idx_quest_archetype_nodes_exposition_template_id
  ON quest_archetype_nodes(exposition_template_id);

CREATE INDEX IF NOT EXISTS idx_tutorial_configs_base_quest_giver_character_template_id
  ON tutorial_configs(base_quest_giver_character_template_id);
