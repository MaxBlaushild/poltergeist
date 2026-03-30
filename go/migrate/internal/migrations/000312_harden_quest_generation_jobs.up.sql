ALTER TABLE quest_generation_jobs
  ADD COLUMN IF NOT EXISTS started_count int NOT NULL DEFAULT 0;

UPDATE quest_generation_jobs
SET started_count = 0
WHERE started_count IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS quest_generation_jobs_active_zone_quest_archetype_idx
  ON quest_generation_jobs(zone_quest_archetype_id)
  WHERE zone_quest_archetype_id IS NOT NULL
    AND status IN ('queued', 'in_progress');
