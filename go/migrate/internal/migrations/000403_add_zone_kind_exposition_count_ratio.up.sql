ALTER TABLE zone_kinds
  ADD COLUMN IF NOT EXISTS exposition_count_ratio DOUBLE PRECISION NOT NULL DEFAULT 1;

UPDATE zone_kinds AS zk
SET exposition_count_ratio = v.exposition_count_ratio
FROM (
  VALUES
    ('academy', 2.4),
    ('badlands', 0.6),
    ('cave', 0.8),
    ('city', 2.2),
    ('coast', 1.2),
    ('desert', 0.7),
    ('farmland', 1.2),
    ('forest', 1.0),
    ('graveyard', 1.3),
    ('highlands', 0.8),
    ('industrial', 1.0),
    ('jungle', 0.9),
    ('mountain', 0.5),
    ('plains', 0.7),
    ('port', 1.8),
    ('reef', 0.9),
    ('riverlands', 1.5),
    ('ruins', 1.5),
    ('sunken-ruins', 1.4),
    ('swamp', 1.1),
    ('temple-grounds', 1.8),
    ('tidal-flats', 1.0),
    ('tundra', 0.5),
    ('village', 1.7),
    ('volcanic', 0.4)
) AS v(slug, exposition_count_ratio)
WHERE zk.slug = v.slug;
