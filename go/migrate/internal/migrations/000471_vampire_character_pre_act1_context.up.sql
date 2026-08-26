-- Pre-Act-1 context becomes mystery-scoped, same as post-Act-1 context
-- already is (see 000462) — the same character cast in two different
-- primary mysteries goes in as someone different in each. Previously this
-- lived only as the character-global PreEventInfo ("Pre-event bio" in the
-- admin UI).
ALTER TABLE vampire_character_mystery_contexts ADD COLUMN IF NOT EXISTS pre_act1_context TEXT NOT NULL DEFAULT '';

-- Backfill: every character gets a pre_act1_context row for every primary
-- mystery (not sub-plots — pre/post-Act-1 context has only ever been
-- scoped to the one required mystery an instance runs, never its
-- subplots), seeded from their current pre_event_info. Existing
-- post_act1_context rows are left untouched; a character with no row yet
-- for a given mystery gets one created with post_act1_context left at its
-- default ''.
INSERT INTO vampire_character_mystery_contexts (character_id, mystery_id, pre_act1_context)
SELECT c.id, m.id, c.pre_event_info
FROM vampire_characters c
CROSS JOIN vampire_mysteries m
WHERE m.is_subplot = false
ON CONFLICT (character_id, mystery_id) DO UPDATE SET pre_act1_context = EXCLUDED.pre_act1_context;
