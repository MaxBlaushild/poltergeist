DELETE FROM zone_kinds
WHERE slug IN (
  'city',
  'village',
  'academy',
  'forest',
  'swamp',
  'badlands',
  'farmland',
  'highlands',
  'mountain',
  'ruins',
  'graveyard',
  'industrial',
  'desert',
  'temple-grounds',
  'volcanic'
);
