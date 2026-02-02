BEGIN;

ALTER TABLE characters
    ADD COLUMN point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL;

CREATE INDEX idx_characters_point_of_interest_id ON characters(point_of_interest_id);

COMMIT;
