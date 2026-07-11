-- Each game can be assigned a GM to run it, plus GM-only "how to run" notes.
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS assigned_gm TEXT NOT NULL DEFAULT '';
ALTER TABLE vampire_games ADD COLUMN IF NOT EXISTS run_notes TEXT NOT NULL DEFAULT '';
