CREATE TABLE IF NOT EXISTS trades_ar_glasses_leads (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  email TEXT NOT NULL UNIQUE,
  trade TEXT NOT NULL DEFAULT '',
  crew_size TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'landing-page',
  user_agent TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS trades_ar_glasses_leads_created_at_idx
  ON trades_ar_glasses_leads(created_at DESC);
