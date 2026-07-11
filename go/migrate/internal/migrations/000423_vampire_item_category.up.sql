-- Free-form category tag for an item (e.g. "War", "Glory", "Protection"),
-- shown to both players and GMs.
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';
