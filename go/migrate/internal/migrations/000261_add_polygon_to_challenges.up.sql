ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS polygon geometry(Polygon,4326);

CREATE INDEX IF NOT EXISTS idx_challenges_polygon
  ON challenges USING GIST(polygon);
