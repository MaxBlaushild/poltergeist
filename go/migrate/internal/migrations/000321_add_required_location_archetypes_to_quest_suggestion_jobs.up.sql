ALTER TABLE quest_archetype_suggestion_jobs
ADD COLUMN IF NOT EXISTS required_location_archetype_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
