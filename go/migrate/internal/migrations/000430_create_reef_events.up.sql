-- Analytics events (R-9.1). Written directly to Postgres per R-9.1's fallback:
-- no existing repo-wide telemetry path exists to CONFORM to (see INVENTORY.md).
CREATE TABLE IF NOT EXISTS reef_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  event_type TEXT NOT NULL,
  session_id TEXT NOT NULL DEFAULT '',
  product_slug TEXT NOT NULL DEFAULT '',
  configuration_id UUID,
  rule TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_reef_events_event_type_created_at ON reef_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reef_events_session_id ON reef_events(session_id);
CREATE INDEX IF NOT EXISTS idx_reef_events_metadata_gin ON reef_events USING GIN (metadata);
