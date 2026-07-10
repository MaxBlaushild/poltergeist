-- Games can be scheduled within the evening (start/end as minutes-of-day, e.g.
-- 6pm = 1080, midnight = 1440) and given a location.
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS start_minutes INTEGER;
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS end_minutes INTEGER;
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS location TEXT NOT NULL DEFAULT '';
