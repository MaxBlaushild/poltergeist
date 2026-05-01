ALTER TABLE quest_archetype_suggestion_drafts
  DROP COLUMN IF EXISTS zone_kind;

ALTER TABLE quest_archetype_suggestion_jobs
  DROP COLUMN IF EXISTS zone_kind;
