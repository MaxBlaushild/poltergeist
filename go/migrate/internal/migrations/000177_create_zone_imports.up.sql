CREATE TABLE IF NOT EXISTS zone_imports (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  metro_name TEXT NOT NULL,
  status TEXT NOT NULL,
  error_message TEXT,
  zone_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_zone_imports_created_at ON zone_imports(created_at DESC);
