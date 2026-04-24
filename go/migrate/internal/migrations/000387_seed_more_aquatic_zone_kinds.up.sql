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
    'Coast',
    'coast',
    'Wind-beaten shorelines with coves, sea caves, cliff paths, wreckage, and a steady mix of scavenging, monsters, and hidden caches.',
    '#4f7f92',
    0.9, 1.1, 1.0, 1.0, 1.0, 1.1, 1.3, 0.8, 0.95, 0.9, 1.0
  ),
  (
    'Riverlands',
    'riverlands',
    'Branching waterways full of ferries, levees, reedbanks, and trade routes that favor travel scenarios and abundant gathering.',
    '#5e8b74',
    1.0, 0.9, 0.8, 0.8, 1.3, 1.4, 0.9, 1.1, 0.95, 1.5, 0.4
  ),
  (
    'Tidal Flats',
    'tidal-flats',
    'Exposed seabeds and shifting inlets where timed access, shell beds, stranded relics, and fast-moving tides reward careful exploration.',
    '#8ca091',
    0.5, 1.0, 0.9, 0.8, 1.2, 1.2, 1.4, 0.4, 1.0, 1.3, 0.7
  ),
  (
    'Sunken Ruins',
    'sunken-ruins',
    'Drowned temples and collapsed causeways thick with lurking threats, trapped relics, and heavy treasure pressure.',
    '#5b7284',
    0.6, 1.3, 1.4, 1.1, 1.0, 1.1, 1.7, 0.4, 0.95, 0.5, 1.4
  ),
  (
    'Port',
    'port',
    'Bustling harbor districts with warehouses, taverns, customs intrigue, smugglers, and conflict around cargo, crews, and contraband.',
    '#6a7486',
    1.6, 0.9, 0.9, 1.2, 1.4, 1.4, 1.0, 0.9, 0.6, 0.2, 1.0
  )
ON CONFLICT (slug) DO NOTHING;
