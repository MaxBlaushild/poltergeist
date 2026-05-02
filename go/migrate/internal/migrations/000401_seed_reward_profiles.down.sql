DELETE FROM reward_profiles
WHERE slug IN (
  'combat',
  'lore',
  'exploration',
  'social',
  'nature',
  'treasure',
  'story'
);
