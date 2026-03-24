ALTER TABLE quest_archetypes
ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS acceptance_dialogue JSONB NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'random',
ADD COLUMN IF NOT EXISTS random_reward_size TEXT NOT NULL DEFAULT 'small',
ADD COLUMN IF NOT EXISTS reward_experience INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS recurrence_frequency TEXT,
ADD COLUMN IF NOT EXISTS material_rewards_json JSONB NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS character_tags JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE IF NOT EXISTS quest_archetype_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  quest_archetype_id UUID NOT NULL REFERENCES quest_archetypes(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_quest_archetype_spell_rewards_quest_archetype_id
  ON quest_archetype_spell_rewards(quest_archetype_id);
