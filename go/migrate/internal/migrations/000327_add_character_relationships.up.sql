CREATE TABLE IF NOT EXISTS user_character_relationships (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id uuid NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    trust INTEGER NOT NULL DEFAULT 0,
    respect INTEGER NOT NULL DEFAULT 0,
    fear INTEGER NOT NULL DEFAULT 0,
    debt INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS user_character_relationships_user_character_idx
    ON user_character_relationships (user_id, character_id);

ALTER TABLE quests
    ADD COLUMN IF NOT EXISTS quest_giver_relationship_effects JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE quest_archetypes
    ADD COLUMN IF NOT EXISTS quest_giver_relationship_effects JSONB NOT NULL DEFAULT '{}'::jsonb;
