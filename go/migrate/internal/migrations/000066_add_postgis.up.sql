CREATE EXTENSION postgis;

ALTER TABLE point_of_interest ADD COLUMN geometry geometry(Point, 4326);
