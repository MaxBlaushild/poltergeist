-- Restore the column, backfilled from bio (its last known content, since
-- the two were identical at the time this migration ran).
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS pre_event_info TEXT NOT NULL DEFAULT '';
UPDATE vampire_characters SET pre_event_info = bio WHERE pre_event_info = '';
