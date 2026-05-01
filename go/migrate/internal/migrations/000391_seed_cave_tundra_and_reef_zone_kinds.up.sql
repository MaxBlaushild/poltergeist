INSERT INTO zone_kinds (
  name,
  slug,
  description,
  overlay_color,
  place_count_ratio,
  monster_count_ratio,
  boss_encounter_count_ratio,
  raid_encounter_count_ratio,
  input_encounter_count_ratio,
  option_encounter_count_ratio,
  treasure_chest_count_ratio,
  healing_fountain_count_ratio,
  resource_count_ratio,
  herbalism_resource_count_ratio,
  mining_resource_count_ratio
)
VALUES
  (
    'Cave',
    'cave',
    'Subterranean tunnels and caverns full of echoing chambers, fungal growth, hidden veins, burrowing threats, and claustrophobic treasure runs.',
    '#6b6f79',
    0.5, 1.3, 1.2, 1.0, 0.9, 0.9, 1.4, 0.4, 1.1, 0.7, 1.8
  ),
  (
    'Tundra',
    'tundra',
    'Frozen open country shaped by bitter wind, sparse shelter, hardy beasts, and survival-focused expeditions across snow and ice.',
    '#8ea7b3',
    0.5, 1.1, 1.2, 1.0, 0.8, 0.9, 1.2, 0.4, 0.75, 0.2, 1.1
  ),
  (
    'Reef',
    'reef',
    'Jagged coral shelves and bright shallows crowded with sea life, wreckage, hidden pearls, and dangerous beauty.',
    '#4ca4a8',
    0.4, 1.2, 1.1, 1.0, 0.9, 1.0, 1.6, 0.6, 1.0, 1.2, 0.8
  )
ON CONFLICT (slug) DO NOTHING;
