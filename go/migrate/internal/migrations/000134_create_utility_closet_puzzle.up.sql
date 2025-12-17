-- Create utility_closet_puzzle table
-- This is a singleton table (only one row should exist)
CREATE TABLE utility_closet_puzzle (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Hue light IDs for each button (0-5)
    button_0_hue_light_id INTEGER,
    button_1_hue_light_id INTEGER,
    button_2_hue_light_id INTEGER,
    button_3_hue_light_id INTEGER,
    button_4_hue_light_id INTEGER,
    button_5_hue_light_id INTEGER,
    -- Current hue state for each button (color index 0-5)
    button_0_current_hue INTEGER NOT NULL DEFAULT 0,
    button_1_current_hue INTEGER NOT NULL DEFAULT 1,
    button_2_current_hue INTEGER NOT NULL DEFAULT 2,
    button_3_current_hue INTEGER NOT NULL DEFAULT 3,
    button_4_current_hue INTEGER NOT NULL DEFAULT 4,
    button_5_current_hue INTEGER NOT NULL DEFAULT 5,
    -- Base hue state for each button (color index 0-5)
    button_0_base_hue INTEGER NOT NULL DEFAULT 0,
    button_1_base_hue INTEGER NOT NULL DEFAULT 1,
    button_2_base_hue INTEGER NOT NULL DEFAULT 2,
    button_3_base_hue INTEGER NOT NULL DEFAULT 3,
    button_4_base_hue INTEGER NOT NULL DEFAULT 4,
    button_5_base_hue INTEGER NOT NULL DEFAULT 5,
    CHECK (button_0_current_hue >= 0 AND button_0_current_hue <= 5),
    CHECK (button_1_current_hue >= 0 AND button_1_current_hue <= 5),
    CHECK (button_2_current_hue >= 0 AND button_2_current_hue <= 5),
    CHECK (button_3_current_hue >= 0 AND button_3_current_hue <= 5),
    CHECK (button_4_current_hue >= 0 AND button_4_current_hue <= 5),
    CHECK (button_5_current_hue >= 0 AND button_5_current_hue <= 5),
    CHECK (button_0_base_hue >= 0 AND button_0_base_hue <= 5),
    CHECK (button_1_base_hue >= 0 AND button_1_base_hue <= 5),
    CHECK (button_2_base_hue >= 0 AND button_2_base_hue <= 5),
    CHECK (button_3_base_hue >= 0 AND button_3_base_hue <= 5),
    CHECK (button_4_base_hue >= 0 AND button_4_base_hue <= 5),
    CHECK (button_5_base_hue >= 0 AND button_5_base_hue <= 5)
);
