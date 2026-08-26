-- PreEventInfo ("Pre-event bio" in the admin UI) is superseded by Bio —
-- its content was backfilled into bio in migration 000470, and every
-- read/write path has moved onto Bio since. Drop the now-unused column.
ALTER TABLE vampire_characters DROP COLUMN IF EXISTS pre_event_info;
