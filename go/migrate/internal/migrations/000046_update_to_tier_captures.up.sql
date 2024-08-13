ALTER TABLE point_of_interest_teams RENAME COLUMN captured TO tier_one_captured;
ALTER TABLE point_of_interest_teams RENAME COLUMN attuned TO tier_two_captured;
ALTER TABLE point_of_interest_teams ADD COLUMN tier_three_captured BOOLEAN;
