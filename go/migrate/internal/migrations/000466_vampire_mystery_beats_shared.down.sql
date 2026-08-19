ALTER TABLE vampire_mystery_beats ADD COLUMN IF NOT EXISTS mystery_id UUID REFERENCES vampire_mysteries(id);
ALTER TABLE vampire_mystery_beats ADD COLUMN IF NOT EXISTS ordinal INT NOT NULL DEFAULT 0;

-- Lossy: a beat shared across multiple mysteries picks one (its
-- lowest-ordinal link) to revert to; the pre-share schema had no way to
-- represent a beat attached to more than one mystery, or to none.
UPDATE vampire_mystery_beats b
SET mystery_id = l.mystery_id, ordinal = l.ordinal
FROM (
  SELECT DISTINCT ON (beat_id) beat_id, mystery_id, ordinal
  FROM vampire_mystery_beat_links
  ORDER BY beat_id, ordinal ASC
) l
WHERE l.beat_id = b.id;

DELETE FROM vampire_mystery_beats WHERE mystery_id IS NULL;

ALTER TABLE vampire_mystery_beats ALTER COLUMN mystery_id SET NOT NULL;

DROP TABLE IF EXISTS vampire_mystery_beat_links;
