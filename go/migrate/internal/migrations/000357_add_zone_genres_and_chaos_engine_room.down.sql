DELETE FROM user_base_structures
WHERE structure_key = 'chaos_engine';

DELETE FROM base_structure_level_visuals
WHERE structure_definition_id IN (
  SELECT id FROM base_structure_definitions WHERE key = 'chaos_engine'
);

DELETE FROM base_structure_level_costs
WHERE structure_definition_id IN (
  SELECT id FROM base_structure_definitions WHERE key = 'chaos_engine'
);

DELETE FROM base_structure_definitions
WHERE key = 'chaos_engine';

DROP TABLE IF EXISTS zone_genre_scores;
DROP TABLE IF EXISTS zone_genres;
