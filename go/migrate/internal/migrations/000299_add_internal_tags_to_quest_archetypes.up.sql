ALTER TABLE quest_archetypes
ADD COLUMN IF NOT EXISTS internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
