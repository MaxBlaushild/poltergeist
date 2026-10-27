-- A lodging photo must never become a hero image merely by losing its scope.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM tc_trip_media WHERE lodging_id IS NOT NULL) THEN
   RAISE EXCEPTION 'Remove lodging photo attachments and their media before rolling back migration 479';
 END IF;
END $$;
DROP INDEX tc_trip_media_lodging_idx;
ALTER TABLE tc_trip_media DROP COLUMN lodging_id;
