-- Shared trip documents are independent of private library bookmarks and of
-- the host's draft/published itinerary. A library import is an explicit copy.
CREATE TABLE tc_trip_documents (
 id uuid PRIMARY KEY,
 stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
 added_by uuid NOT NULL REFERENCES users(id),
 title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 500),
 kind text NOT NULL CHECK (kind IN ('file', 'link', 'guide')),
 format text NOT NULL DEFAULT 'guide',
 link text NOT NULL DEFAULT '',
 content text NOT NULL DEFAULT '',
 filename text NOT NULL DEFAULT '',
 content_type text NOT NULL DEFAULT '',
 size_bytes bigint NOT NULL DEFAULT 0 CHECK (size_bytes BETWEEN 0 AND 20971520),
 storage_bucket text NOT NULL DEFAULT '',
 storage_key text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE (stop_id, id),
 CHECK ((kind = 'file' AND storage_bucket <> '' AND storage_key <> '' AND size_bytes > 0)
     OR (kind <> 'file' AND storage_bucket = '' AND storage_key = '' AND size_bytes = 0))
);
CREATE INDEX tc_trip_documents_stop_idx ON tc_trip_documents(stop_id, created_at DESC, id DESC);

CREATE TABLE tc_trip_document_comments (
 id uuid PRIMARY KEY,
 stop_id uuid NOT NULL,
 document_id uuid NOT NULL,
 author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY (stop_id, document_id) REFERENCES tc_trip_documents(stop_id, id) ON DELETE CASCADE
);
CREATE INDEX tc_trip_document_comments_document_idx ON tc_trip_document_comments(document_id, created_at DESC, id DESC);
