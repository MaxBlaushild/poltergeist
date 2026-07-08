-- House taglines ("Order is power", etc.) and per-character portrait images.
ALTER TABLE vampire_houses ADD COLUMN IF NOT EXISTS tagline TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
