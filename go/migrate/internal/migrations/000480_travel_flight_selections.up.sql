CREATE TABLE tc_flight_selections (
 stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
 flight_id uuid NOT NULL,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 status text NOT NULL CHECK (status IN ('considering','booked','removed')),
 leg_ids jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(leg_ids) = 'array'),
 leg_versions jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(leg_versions) = 'object'),
 share_with_travelers boolean NOT NULL DEFAULT false,
 roster_consent_version integer NOT NULL DEFAULT 0,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY (stop_id, flight_id, user_id)
);
CREATE INDEX tc_flight_selections_user_idx ON tc_flight_selections(user_id, stop_id);
