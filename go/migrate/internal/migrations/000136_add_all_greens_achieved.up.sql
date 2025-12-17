-- Add all_greens_achieved field to utility_closet_puzzle table
-- This field tracks whether all lights have been green at least once,
-- which is required to unlock the blue color from red

ALTER TABLE utility_closet_puzzle
ADD COLUMN all_greens_achieved BOOLEAN NOT NULL DEFAULT false;
