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
    'Plains',
    'plains',
    'Wide grasslands with open sightlines, roaming herds, wind-scoured travel, and sparse shelter that favors pursuit and exposed encounters.',
    '#9bad62',
    0.7, 1.0, 1.0, 1.1, 0.9, 1.0, 0.8, 0.9, 0.8, 1.3, 0.3
  ),
  (
    'Jungle',
    'jungle',
    'Dense tropical overgrowth with tangled canopy, hidden ruins, venomous wildlife, and high-pressure exploration rich in herbs and ambushes.',
    '#3f7a4e',
    0.6, 1.4, 1.2, 1.1, 1.0, 1.1, 1.3, 0.7, 1.1, 1.8, 0.4
  )
ON CONFLICT (slug) DO NOTHING;
