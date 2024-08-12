ALTER TABLE points_of_interest RENAME COLUMN capture_challenge TO tier_one_challenge;
ALTER TABLE points_of_interest RENAME COLUMN attune_challenge TO tier_two_challenge;
ALTER TABLE points_of_interest ADD COLUMN tier_three_challenge TEXT;
