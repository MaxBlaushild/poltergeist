-- Structured Blood Token effects for items, resolved into the final tally at the
-- Final Reveal. Free-text `effect` stays for display; these drive computation.
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS bt_self INTEGER NOT NULL DEFAULT 0;           -- flat BT to the owner
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS bt_from_target INTEGER NOT NULL DEFAULT 0;    -- steal N: +N owner, -N target
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS bt_deduct_target INTEGER NOT NULL DEFAULT 0;  -- deduct N from target (no gain)
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS quiz_bt_pct INTEGER NOT NULL DEFAULT 0;       -- +pct% of owner's Part 1 quiz BT
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS double_game_bt BOOLEAN NOT NULL DEFAULT FALSE;-- add another copy of owner's game BT
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS immune BOOLEAN NOT NULL DEFAULT FALSE;        -- cancel incoming steals/deducts
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS reflect BOOLEAN NOT NULL DEFAULT FALSE;       -- bounce incoming loss to the attacker
ALTER TABLE vampire_items ADD COLUMN IF NOT EXISTS strip_resistance BOOLEAN NOT NULL DEFAULT FALSE; -- ignore target's immune/reflect
