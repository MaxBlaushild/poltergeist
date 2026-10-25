ALTER TABLE tc_stops
 ADD COLUMN host_model boolean NOT NULL DEFAULT false,
 ADD COLUMN revision integer NOT NULL DEFAULT 1 CHECK (revision > 0),
 ADD COLUMN max_attendees integer CHECK (max_attendees IS NULL OR max_attendees > 0),
 ADD COLUMN waitlist_enabled boolean NOT NULL DEFAULT false,
 ADD COLUMN roster_visibility text NOT NULL DEFAULT 'attendees' CHECK (roster_visibility IN ('attendees','hosts')),
 ADD COLUMN show_attendee_count boolean NOT NULL DEFAULT true,
 ADD COLUMN roster_version integer NOT NULL DEFAULT 1 CHECK (roster_version > 0);
ALTER TABLE tc_collaborators ADD COLUMN is_host boolean NOT NULL DEFAULT false;
ALTER TABLE tc_participations
 ADD COLUMN reserved boolean NOT NULL DEFAULT false,
 ADD COLUMN roster_consent_version integer NOT NULL DEFAULT 0,
 ADD COLUMN share_dates boolean NOT NULL DEFAULT false;
ALTER TABLE tc_participations DROP CONSTRAINT tc_participations_status_check;
ALTER TABLE tc_participations ADD CONSTRAINT tc_participations_status_check
 CHECK (status IN ('interested','requested','joined','needs_reconfirmation','declined','withdrawn','removed','cancelled','waitlisted'));
ALTER TABLE tc_participations ADD CONSTRAINT tc_participations_reserved_status_check
 CHECK (NOT reserved OR status IN ('requested','needs_reconfirmation'));
CREATE TABLE tc_trip_audit_events (
 id uuid PRIMARY KEY, stop_id uuid NOT NULL, actor_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 action text NOT NULL, details jsonb NOT NULL DEFAULT '{}', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_trip_audit_stop_idx ON tc_trip_audit_events(stop_id,created_at);
CREATE TABLE tc_trip_media (
 id uuid PRIMARY KEY, stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
 uploaded_by uuid NOT NULL REFERENCES users(id), content_type text NOT NULL,
 width integer NOT NULL CHECK(width>0), height integer NOT NULL CHECK(height>0),
 data bytea NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_trip_media_stop_idx ON tc_trip_media(stop_id);
