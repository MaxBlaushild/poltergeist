ALTER TABLE points_of_interest
  ADD COLUMN IF NOT EXISTS marker_category VARCHAR(64);
