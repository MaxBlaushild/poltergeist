ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS fetch_character_id UUID,
  ADD COLUMN IF NOT EXISTS fetch_requirements_json JSONB NOT NULL DEFAULT '[]';

ALTER TABLE quest_nodes
  ADD COLUMN IF NOT EXISTS fetch_character_id UUID,
  ADD COLUMN IF NOT EXISTS fetch_requirements_json JSONB NOT NULL DEFAULT '[]';
