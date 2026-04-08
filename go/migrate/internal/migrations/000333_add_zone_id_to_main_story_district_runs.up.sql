ALTER TABLE main_story_district_runs
    ADD COLUMN IF NOT EXISTS zone_id uuid REFERENCES zones(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_main_story_district_runs_zone_id
    ON main_story_district_runs(zone_id);
