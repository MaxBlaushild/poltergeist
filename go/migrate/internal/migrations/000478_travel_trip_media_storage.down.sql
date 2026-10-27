-- Preserve images on rollback: copy external objects back into data and clear
-- their storage columns before removing this migration.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM tc_trip_media WHERE storage_key <> '') THEN
   RAISE EXCEPTION 'Restore S3 trip media bytes to tc_trip_media.data before rolling back migration 478';
 END IF;
END $$;
ALTER TABLE tc_trip_media
 DROP CONSTRAINT tc_trip_media_storage_check,
 DROP COLUMN storage_key,
 DROP COLUMN storage_bucket;
