-- Beats become shared, reusable content instead of belonging to exactly
-- one mystery: the same beat (e.g. "the tools to kill a vampire") can now
-- be attached to multiple mysteries/subplots at once, each with its own
-- position in that mystery's list. Editing a shared beat's title/
-- description changes it everywhere it's linked; unlinking it from one
-- mystery only removes that link, never the beat itself.
CREATE TABLE IF NOT EXISTS vampire_mystery_beat_links (
  mystery_id UUID NOT NULL REFERENCES vampire_mysteries(id) ON DELETE CASCADE,
  beat_id UUID NOT NULL REFERENCES vampire_mystery_beats(id) ON DELETE CASCADE,
  ordinal INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (mystery_id, beat_id)
);
CREATE INDEX IF NOT EXISTS idx_vampire_mystery_beat_links_beat_id ON vampire_mystery_beat_links(beat_id);

-- Backfill: every beat's existing (mystery_id, ordinal) becomes a link row.
INSERT INTO vampire_mystery_beat_links (mystery_id, beat_id, ordinal)
SELECT mystery_id, id, ordinal FROM vampire_mystery_beats
ON CONFLICT (mystery_id, beat_id) DO NOTHING;

ALTER TABLE vampire_mystery_beats DROP COLUMN IF EXISTS mystery_id;
ALTER TABLE vampire_mystery_beats DROP COLUMN IF EXISTS ordinal;
