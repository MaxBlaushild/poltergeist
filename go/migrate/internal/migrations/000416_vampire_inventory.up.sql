-- Inventory: a catalog of assignable items, and per-player ownership.
CREATE TABLE IF NOT EXISTS vampire_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    code TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    effect TEXT NOT NULL DEFAULT '',
    -- Whether the effect targets another player (shows a target picker in the app).
    targets_player BOOLEAN NOT NULL DEFAULT FALSE,
    -- House Favor auto-applied to the owner's house at the Final Reveal (0 = none).
    hf_effect INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS vampire_player_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    player_id UUID NOT NULL REFERENCES vampire_players(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES vampire_items(id) ON DELETE CASCADE,
    -- For targeting items: which player this instance is aimed at.
    target_player_id UUID REFERENCES vampire_players(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_vampire_player_items_player ON vampire_player_items(player_id);
