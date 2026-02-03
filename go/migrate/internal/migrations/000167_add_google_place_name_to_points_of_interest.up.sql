ALTER TABLE points_of_interest
ADD COLUMN google_maps_place_name TEXT;

CREATE INDEX IF NOT EXISTS points_of_interest_google_place_name_idx
ON points_of_interest (google_maps_place_name);
