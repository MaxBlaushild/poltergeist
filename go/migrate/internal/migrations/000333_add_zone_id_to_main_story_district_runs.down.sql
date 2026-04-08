DROP INDEX IF EXISTS idx_main_story_district_runs_zone_id;

ALTER TABLE main_story_district_runs
    DROP COLUMN IF EXISTS zone_id;
