ALTER TABLE points_of_interest RENAME COLUMN tier_one_challenge TO capture_challenge;
ALTER TABLE points_of_interest RENAME COLUMN tier_two_challenge TO attune_challenge;
ALTER TABLE points_of_interest DROP COLUMN tier_three_challenge;

