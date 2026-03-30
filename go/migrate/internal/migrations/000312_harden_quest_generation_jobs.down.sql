DROP INDEX IF EXISTS quest_generation_jobs_active_zone_quest_archetype_idx;

ALTER TABLE quest_generation_jobs
  DROP COLUMN IF EXISTS started_count;
