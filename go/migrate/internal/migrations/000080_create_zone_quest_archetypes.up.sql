CREATE TABLE zone_quest_archetypes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    zone_id UUID NOT NULL REFERENCES zones(id),
    number_of_quests INTEGER NOT NULL,
    quest_archetype_id UUID NOT NULL REFERENCES quest_archetypes(id)
);
