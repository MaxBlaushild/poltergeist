-- Add all_purples_achieved field to utility_closet_puzzle table
-- This field tracks whether all lights have been purple at least once,
-- which is required to unlock the white color

ALTER TABLE utility_closet_puzzle
ADD COLUMN all_purples_achieved BOOLEAN NOT NULL DEFAULT false;
