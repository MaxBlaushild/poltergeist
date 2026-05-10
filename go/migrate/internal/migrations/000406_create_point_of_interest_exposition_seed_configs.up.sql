CREATE TABLE IF NOT EXISTS point_of_interest_exposition_seed_configs (
  id INTEGER PRIMARY KEY,
  profiles_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO point_of_interest_exposition_seed_configs (id, profiles_json)
VALUES (1, '[]')
ON CONFLICT (id) DO NOTHING;
