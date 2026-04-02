ALTER TABLE quest_archetypes
ADD COLUMN IF NOT EXISTS quest_giver_character_id uuid REFERENCES characters(id);

CREATE INDEX IF NOT EXISTS idx_quest_archetypes_quest_giver_character_id
ON quest_archetypes (quest_giver_character_id);
