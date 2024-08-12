ALTER TABLE points_of_interest RENAME COLUMN tier_one_captured TO captured;
ALTER TABLE points_of_interest RENAME COLUMN tier_two_captured TO attuned;
ALTER TABLE points_of_interest DROP COLUMN tier_three_captured;

