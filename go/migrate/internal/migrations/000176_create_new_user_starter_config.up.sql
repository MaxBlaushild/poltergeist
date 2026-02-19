CREATE TABLE IF NOT EXISTS new_user_starter_configs (
  id INTEGER PRIMARY KEY,
  gold INTEGER NOT NULL DEFAULT 0,
  items_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS new_user_starter_grants (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO new_user_starter_configs (id, gold, items_json)
VALUES (1, 0, '[]')
ON CONFLICT (id) DO NOTHING;
