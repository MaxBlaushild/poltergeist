BEGIN;

ALTER TABLE point_of_interest_groups
    DROP COLUMN IF EXISTS required_reputation_level;

COMMIT;
