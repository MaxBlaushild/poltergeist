BEGIN;

CREATE TABLE character_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    dialogue JSONB,
    metadata JSONB
);

CREATE INDEX idx_character_actions_character_id ON character_actions(character_id);
CREATE INDEX idx_character_actions_action_type ON character_actions(action_type);

COMMIT;

