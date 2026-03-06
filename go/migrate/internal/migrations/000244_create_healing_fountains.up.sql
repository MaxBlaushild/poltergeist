CREATE TABLE healing_fountains (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL DEFAULT 'Healing Fountain',
  description TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326),
  invalidated BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_healing_fountains_zone_id ON healing_fountains(zone_id);
CREATE INDEX idx_healing_fountains_geometry ON healing_fountains USING GIST(geometry);

CREATE TABLE user_healing_fountain_visits (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  healing_fountain_id UUID NOT NULL REFERENCES healing_fountains(id) ON DELETE CASCADE,
  visited_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_healing_fountain_visits_user_id
  ON user_healing_fountain_visits(user_id);
CREATE INDEX idx_user_healing_fountain_visits_fountain_id
  ON user_healing_fountain_visits(healing_fountain_id);
CREATE INDEX idx_user_healing_fountain_visits_user_visited_at
  ON user_healing_fountain_visits(user_id, visited_at DESC);
CREATE INDEX idx_user_healing_fountain_visits_user_fountain_visited_at
  ON user_healing_fountain_visits(user_id, healing_fountain_id, visited_at DESC);
