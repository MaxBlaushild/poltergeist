CREATE TABLE IF NOT EXISTS monster_template_progressions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    monster_template_id UUID NOT NULL REFERENCES monster_templates(id) ON DELETE CASCADE,
    progression_id UUID NOT NULL REFERENCES spell_progressions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_monster_template_progressions_template_progression
    ON monster_template_progressions (monster_template_id, progression_id);

CREATE INDEX IF NOT EXISTS idx_monster_template_progressions_progression_id
    ON monster_template_progressions (progression_id);

INSERT INTO monster_template_progressions (
    id,
    created_at,
    updated_at,
    monster_template_id,
    progression_id
)
SELECT DISTINCT
    uuid_generate_v4(),
    NOW(),
    NOW(),
    mts.monster_template_id,
    sps.progression_id
FROM monster_template_spells mts
JOIN spell_progression_spells sps ON sps.spell_id = mts.spell_id
ON CONFLICT (monster_template_id, progression_id) DO NOTHING;
