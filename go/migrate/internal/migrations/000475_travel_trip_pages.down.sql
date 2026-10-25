DROP TABLE IF EXISTS tc_trip_media;
DROP TABLE IF EXISTS tc_trip_audit_events;
ALTER TABLE tc_participations DROP CONSTRAINT tc_participations_reserved_status_check;
UPDATE tc_participations SET status = 'withdrawn' WHERE status = 'waitlisted';
ALTER TABLE tc_participations DROP CONSTRAINT tc_participations_status_check;
ALTER TABLE tc_participations ADD CONSTRAINT tc_participations_status_check CHECK (status IN ('interested','requested','joined','needs_reconfirmation','declined','withdrawn','removed','cancelled'));
ALTER TABLE tc_participations DROP COLUMN reserved, DROP COLUMN roster_consent_version, DROP COLUMN share_dates;
ALTER TABLE tc_collaborators DROP COLUMN is_host;
ALTER TABLE tc_stops DROP COLUMN host_model, DROP COLUMN revision, DROP COLUMN max_attendees, DROP COLUMN waitlist_enabled, DROP COLUMN roster_visibility, DROP COLUMN show_attendee_count, DROP COLUMN roster_version;
