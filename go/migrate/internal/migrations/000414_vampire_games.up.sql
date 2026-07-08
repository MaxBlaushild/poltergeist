-- Physical games: a pre-seeded (or GM-added) list of the night's contests. When a
-- game is scored, its top-three finishers are recorded here for display and the
-- Blood Token / House Favor awards are written to the ledgers.
CREATE TABLE IF NOT EXISTS vampire_games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ordinal INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending', -- pending | played
    first_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL,
    second_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL,
    third_character_id UUID REFERENCES vampire_characters(id) ON DELETE SET NULL
);
