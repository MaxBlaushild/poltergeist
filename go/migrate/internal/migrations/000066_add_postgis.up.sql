CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE points_of_interest ADD COLUMN geometry geometry(Point, 4326);
