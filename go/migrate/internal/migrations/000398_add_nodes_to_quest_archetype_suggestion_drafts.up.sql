ALTER TABLE quest_archetype_suggestion_drafts
ADD COLUMN nodes JSONB NOT NULL DEFAULT '[]'::jsonb;
