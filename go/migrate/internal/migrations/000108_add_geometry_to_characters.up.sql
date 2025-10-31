BEGIN;

-- Add geometry column to characters table
ALTER TABLE characters ADD COLUMN geometry geometry(Point, 4326);

-- Populate geometry for existing characters using their movement pattern's starting position
UPDATE characters
SET geometry = ST_SetSRID(ST_MakePoint(
    movement_patterns.starting_longitude,
    movement_patterns.starting_latitude
), 4326)
FROM movement_patterns
WHERE characters.movement_pattern_id = movement_patterns.id
    AND movement_patterns.starting_longitude IS NOT NULL
    AND movement_patterns.starting_latitude IS NOT NULL;

-- Remove location_id column and its index
DROP INDEX IF EXISTS idx_characters_location_id;
ALTER TABLE characters DROP COLUMN IF EXISTS location_id;

COMMIT;
