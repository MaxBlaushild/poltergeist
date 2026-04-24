DELETE FROM zone_kinds
WHERE slug IN (
  'coast',
  'riverlands',
  'tidal-flats',
  'sunken-ruins',
  'port'
);
