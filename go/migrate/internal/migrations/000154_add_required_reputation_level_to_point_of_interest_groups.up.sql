BEGIN;

ALTER TABLE point_of_interest_groups
    ADD COLUMN required_reputation_level INT;

COMMIT;
