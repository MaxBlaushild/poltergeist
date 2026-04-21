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
    'City',
    'city',
    'Dense urban streets with social intrigue, dense points of interest, and more choice-driven scenarios than raw gathering.',
    '#6c7a89',
    1.5, 0.9, 0.8, 0.9, 1.4, 1.5, 0.8, 0.9, 0.5, 0.4, 0.6
  ),
  (
    'Village',
    'village',
    'Cozy settlements with local personalities, accessible quests, and light foraging around the edges.',
    '#b59a63',
    1.3, 0.8, 0.7, 0.7, 1.2, 1.3, 0.9, 1.1, 0.85, 1.2, 0.5
  ),
  (
    'Academy',
    'academy',
    'Scholarly grounds with magical study, lore-driven scenarios, and a calmer combat profile.',
    '#7c6fa1',
    1.3, 0.7, 0.8, 0.7, 1.6, 1.5, 1.0, 1.3, 0.65, 0.8, 0.5
  ),
  (
    'Forest',
    'forest',
    'Wild woodland rich in herbs, trails, shrines, and beast pressure.',
    '#567a4e',
    0.8, 1.2, 1.0, 0.9, 0.9, 1.0, 0.9, 1.2, 1.1, 1.7, 0.5
  ),
  (
    'Swamp',
    'swamp',
    'Misty wetlands full of strange growth, unsettling encounters, and potent gatherables.',
    '#556b5d',
    0.7, 1.2, 1.1, 0.9, 1.1, 1.2, 1.0, 0.8, 1.0, 1.6, 0.4
  ),
  (
    'Badlands',
    'badlands',
    'Harsh exposed frontier with sparse shelter, dangerous fights, and mineral-heavy scavenging.',
    '#a06f4f',
    0.6, 1.3, 1.4, 1.2, 0.8, 0.9, 1.2, 0.5, 0.9, 0.3, 1.5
  ),
  (
    'Farmland',
    'farmland',
    'Cultivated rural space with approachable encounters, local tasks, and abundant natural gathering.',
    '#96a85b',
    1.1, 0.8, 0.7, 0.7, 1.0, 1.1, 1.0, 1.0, 0.95, 1.5, 0.4
  ),
  (
    'Highlands',
    'highlands',
    'Windy uplands balancing rugged travel, elevated danger, and a mixed herb-and-ore economy.',
    '#7d8a6a',
    0.8, 1.2, 1.3, 1.2, 0.8, 1.0, 1.1, 0.8, 1.1, 0.9, 1.3
  ),
  (
    'Mountain',
    'mountain',
    'Steep, hostile terrain with elite threats, thin recovery, and strong mining identity.',
    '#7b7f87',
    0.6, 1.3, 1.4, 1.3, 0.7, 0.9, 1.2, 0.7, 1.1, 0.4, 1.8
  ),
  (
    'Ruins',
    'ruins',
    'Ancient remnants full of hidden rewards, lurking danger, and relic-rich exploration.',
    '#8a6a4e',
    0.7, 1.2, 1.3, 1.1, 1.0, 1.2, 1.5, 0.7, 1.05, 0.7, 1.4
  ),
  (
    'Graveyard',
    'graveyard',
    'Haunted ground with oppressive atmosphere, frequent enemies, and low natural recovery.',
    '#5d6675',
    0.6, 1.3, 1.2, 0.9, 1.0, 1.1, 1.0, 0.6, 0.75, 0.6, 0.9
  ),
  (
    'Industrial',
    'industrial',
    'Factories, rail spurs, and hard edges with conflict-heavy pacing and plenty of metal.',
    '#7b6454',
    1.1, 1.1, 1.0, 1.3, 1.1, 1.0, 0.9, 0.6, 0.9, 0.3, 1.5
  ),
  (
    'Desert',
    'desert',
    'Sparse, punishing expanses with hidden caches, low healing, and strong extraction themes.',
    '#c49a5a',
    0.6, 1.1, 1.2, 1.0, 0.9, 1.0, 1.3, 0.5, 0.75, 0.2, 1.3
  ),
  (
    'Temple Grounds',
    'temple-grounds',
    'Sacred precincts built around reflection, ritual, restorative spaces, and spiritual tasks.',
    '#a28b62',
    0.9, 0.8, 1.0, 0.8, 1.4, 1.3, 1.0, 1.6, 0.75, 1.0, 0.5
  ),
  (
    'Volcanic',
    'volcanic',
    'Molten terrain with extreme danger, low forgiveness, rare rewards, and exceptional mineral density.',
    '#943f2f',
    0.5, 1.4, 1.5, 1.3, 0.9, 0.8, 1.4, 0.4, 0.9, 0.1, 1.7
  )
ON CONFLICT (slug) DO NOTHING;
