ALTER TABLE zone_quest_archetypes
ADD COLUMN character_id UUID REFERENCES characters(id);
