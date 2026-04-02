CREATE TABLE IF NOT EXISTS story_world_changes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    main_story_template_id UUID NOT NULL REFERENCES main_story_templates(id) ON DELETE CASCADE,
    quest_archetype_id UUID NULL REFERENCES quest_archetypes(id) ON DELETE SET NULL,
    beat_order INTEGER NOT NULL DEFAULT 0,
    priority INTEGER NOT NULL DEFAULT 0,
    effect_type TEXT NOT NULL,
    target_key TEXT NOT NULL DEFAULT '',
    required_story_flags JSONB NOT NULL DEFAULT '[]'::jsonb,
    character_id UUID NULL REFERENCES characters(id) ON DELETE CASCADE,
    point_of_interest_id UUID NULL REFERENCES points_of_interest(id) ON DELETE CASCADE,
    destination_point_of_interest_id UUID NULL REFERENCES points_of_interest(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    clue TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_story_world_changes_template_id
    ON story_world_changes (main_story_template_id);

CREATE INDEX IF NOT EXISTS idx_story_world_changes_effect_type
    ON story_world_changes (effect_type);

CREATE INDEX IF NOT EXISTS idx_story_world_changes_character_id
    ON story_world_changes (character_id);

CREATE INDEX IF NOT EXISTS idx_story_world_changes_point_of_interest_id
    ON story_world_changes (point_of_interest_id);
