-- Allow gold color (value 6) in utility closet puzzle
-- Update constraints to allow hue values 0-6 instead of 0-5

ALTER TABLE utility_closet_puzzle
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_0_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_1_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_2_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_3_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_4_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_5_current_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_0_base_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_1_base_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_2_base_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_3_base_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_4_base_hue_check,
DROP CONSTRAINT IF EXISTS utility_closet_puzzle_button_5_base_hue_check;

ALTER TABLE utility_closet_puzzle
ADD CONSTRAINT utility_closet_puzzle_button_0_current_hue_check CHECK (button_0_current_hue >= 0 AND button_0_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_1_current_hue_check CHECK (button_1_current_hue >= 0 AND button_1_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_2_current_hue_check CHECK (button_2_current_hue >= 0 AND button_2_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_3_current_hue_check CHECK (button_3_current_hue >= 0 AND button_3_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_4_current_hue_check CHECK (button_4_current_hue >= 0 AND button_4_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_5_current_hue_check CHECK (button_5_current_hue >= 0 AND button_5_current_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_0_base_hue_check CHECK (button_0_base_hue >= 0 AND button_0_base_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_1_base_hue_check CHECK (button_1_base_hue >= 0 AND button_1_base_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_2_base_hue_check CHECK (button_2_base_hue >= 0 AND button_2_base_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_3_base_hue_check CHECK (button_3_base_hue >= 0 AND button_3_base_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_4_base_hue_check CHECK (button_4_base_hue >= 0 AND button_4_base_hue <= 6),
ADD CONSTRAINT utility_closet_puzzle_button_5_base_hue_check CHECK (button_5_base_hue >= 0 AND button_5_base_hue <= 6);
