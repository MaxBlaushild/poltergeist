ALTER TABLE quest_archetype_suggestion_jobs
ADD COLUMN IF NOT EXISTS family_mix_targets JSONB NOT NULL DEFAULT '{}'::jsonb;
