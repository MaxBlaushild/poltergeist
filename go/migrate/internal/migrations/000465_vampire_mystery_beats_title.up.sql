-- Beats gain a title, separate from their body text. Renaming body ->
-- description for clarity now that there are two fields — a short title
-- (list/dropdown display) and the longer description (what actually gets
-- revealed).
ALTER TABLE vampire_mystery_beats RENAME COLUMN body TO description;
ALTER TABLE vampire_mystery_beats ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
