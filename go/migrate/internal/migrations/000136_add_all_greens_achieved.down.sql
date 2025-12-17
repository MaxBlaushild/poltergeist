-- Remove all_greens_achieved field from utility_closet_puzzle table
ALTER TABLE utility_closet_puzzle
DROP COLUMN IF EXISTS all_greens_achieved;
