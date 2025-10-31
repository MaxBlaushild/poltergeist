BEGIN;

-- Re-add location_id column
ALTER TABLE characters ADD COLUMN location_id UUID REFERENCES points_of_interest(id);

-- Recreate index
CREATE INDEX idx_characters_location_id ON characters(location_id);

-- Drop geometry column
ALTER TABLE characters DROP COLUMN IF EXISTS geometry;

COMMIT;
