ALTER TABLE documents ADD COLUMN visibility text NOT NULL DEFAULT 'network';
ALTER TABLE documents ALTER COLUMN visibility SET DEFAULT 'private';
ALTER TABLE documents ADD CONSTRAINT documents_visibility_check CHECK (visibility IN ('private', 'network', 'selected'));
ALTER TABLE documents ADD COLUMN format text NOT NULL DEFAULT 'guide';
UPDATE documents SET format = 'spreadsheet' WHERE provider = 'google_sheets';
UPDATE documents SET format = 'video' WHERE lower(COALESCE(link, '')) ~ '(/videos/|\.(mp4|mov|m4v)(\?|$))';
ALTER TABLE documents ADD CONSTRAINT documents_format_check CHECK (format IN ('spreadsheet', 'itinerary', 'guide', 'video'));
ALTER TABLE documents ADD COLUMN personal_note text NOT NULL DEFAULT '';
ALTER TABLE documents ADD COLUMN destination text NOT NULL DEFAULT '';
CREATE INDEX documents_library_audience_idx ON documents (visibility, user_id, created_at DESC);
CREATE TABLE document_shares (
  document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(document_id, user_id)
);
CREATE INDEX document_shares_user_idx ON document_shares(user_id, document_id);
CREATE TABLE document_saves (
  document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(document_id, user_id)
);
CREATE INDEX document_saves_user_idx ON document_saves(user_id, document_id);
CREATE TABLE document_trip_references (
  document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  trip_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(document_id, user_id, trip_id)
);
CREATE INDEX document_trip_references_user_idx ON document_trip_references(user_id, document_id);
-- Historical global aggregates cannot serve audience-filtered discovery.
DELETE FROM trending_destinations;
