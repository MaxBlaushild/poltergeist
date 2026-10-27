ALTER TABLE tc_trip_media ADD COLUMN lodging_id uuid;
CREATE INDEX tc_trip_media_lodging_idx ON tc_trip_media(stop_id, lodging_id) WHERE lodging_id IS NOT NULL;
