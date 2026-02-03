CREATE TABLE IF NOT EXISTS point_of_interest_imports (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  place_id TEXT NOT NULL,
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  error_message TEXT,
  point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_point_of_interest_imports_zone_id ON point_of_interest_imports(zone_id);
CREATE INDEX IF NOT EXISTS idx_point_of_interest_imports_status ON point_of_interest_imports(status);
