DELETE FROM quest_generation_jobs
WHERE zone_quest_archetype_id IS NULL;

ALTER TABLE quest_generation_jobs
  ALTER COLUMN zone_quest_archetype_id SET NOT NULL;
