ALTER TABLE quest_archetype_suggestion_jobs
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';

ALTER TABLE quest_archetype_suggestion_drafts
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';
