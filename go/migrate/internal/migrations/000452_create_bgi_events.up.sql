-- Mirrors reef_events (000430). fit_indicator_shown (R-3.4/R-5.3/R-6.3) is
-- bgi's own new event type: every time the configurator computes assembled
-- height vs. box depth, client or server side, this is logged so rejection
-- frequency can reveal whether the seeded box/sleeve numbers need
-- correcting before customers find out physically.
CREATE TABLE IF NOT EXISTS bgi_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  event_type TEXT NOT NULL,
  session_id TEXT NOT NULL DEFAULT '',
  game_slug TEXT NOT NULL DEFAULT '',
  configuration_id UUID,
  rule TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_bgi_events_event_type_created_at ON bgi_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bgi_events_session_id ON bgi_events(session_id);
CREATE INDEX IF NOT EXISTS idx_bgi_events_metadata_gin ON bgi_events USING GIN (metadata);
