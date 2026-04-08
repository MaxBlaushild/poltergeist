ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS base_quest_archetype_id UUID REFERENCES quest_archetypes(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS base_quest_giver_character_id UUID REFERENCES characters(id) ON DELETE SET NULL;

ALTER TABLE characters
  ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS ephemeral BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_characters_owner_user_id ON characters(owner_user_id);

ALTER TABLE quests
  ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS ephemeral BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_quests_owner_user_id ON quests(owner_user_id);
