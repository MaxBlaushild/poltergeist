ALTER TABLE point_of_interest_groups
ADD COLUMN quest_giver_character_id UUID REFERENCES characters(id);

