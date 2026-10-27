ALTER TABLE tc_trip_media
 ADD COLUMN storage_bucket text NOT NULL DEFAULT '',
 ADD COLUMN storage_key text NOT NULL DEFAULT '',
 ADD CONSTRAINT tc_trip_media_storage_check CHECK (
   (storage_bucket = '' AND storage_key = '' AND octet_length(data) > 0)
   OR (storage_bucket <> '' AND storage_key <> '' AND octet_length(data) = 0)
 );
