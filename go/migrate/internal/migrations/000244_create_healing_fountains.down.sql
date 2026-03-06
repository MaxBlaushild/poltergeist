DROP INDEX IF EXISTS idx_user_healing_fountain_visits_user_fountain_visited_at;
DROP INDEX IF EXISTS idx_user_healing_fountain_visits_user_visited_at;
DROP INDEX IF EXISTS idx_user_healing_fountain_visits_fountain_id;
DROP INDEX IF EXISTS idx_user_healing_fountain_visits_user_id;
DROP TABLE IF EXISTS user_healing_fountain_visits;

DROP INDEX IF EXISTS idx_healing_fountains_geometry;
DROP INDEX IF EXISTS idx_healing_fountains_zone_id;
DROP TABLE IF EXISTS healing_fountains;
