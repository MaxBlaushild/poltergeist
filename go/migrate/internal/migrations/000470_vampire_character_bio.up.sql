-- A character's bio, distinct from PreEventInfo ("Pre-event bio" in the
-- admin UI) — that field predates mystery-scoping and doubles as GM/player
-- prep material shown across every read surface. Bio is a separate slot for
-- a character's actual biography, editable independently going forward.
-- Backfilled from the current pre_event_info so existing content isn't
-- lost; the two are free to diverge from here.
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';
UPDATE vampire_characters SET bio = pre_event_info WHERE bio = '';
