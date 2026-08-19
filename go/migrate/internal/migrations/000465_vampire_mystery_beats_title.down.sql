ALTER TABLE vampire_mystery_beats DROP COLUMN IF EXISTS title;
ALTER TABLE vampire_mystery_beats RENAME COLUMN description TO body;
