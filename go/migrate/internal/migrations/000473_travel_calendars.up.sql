CREATE TABLE tc_calendars (
 id uuid PRIMARY KEY, owner_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
 title text NOT NULL, visibility text NOT NULL DEFAULT 'private' CHECK (visibility IN ('private','specific','link')),
 share_token text NOT NULL UNIQUE, sharing_enabled boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE tc_stops (
 id uuid PRIMARY KEY, calendar_id uuid NOT NULL REFERENCES tc_calendars(id) ON DELETE CASCADE,
 draft jsonb NOT NULL, published jsonb,
 visibility text NOT NULL DEFAULT 'inherit' CHECK (visibility IN ('inherit','private','specific','link')),
 share_token text NOT NULL UNIQUE, sharing_enabled boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_stops_calendar_idx ON tc_stops(calendar_id);
CREATE TABLE tc_grants (
 id uuid PRIMARY KEY, calendar_id uuid REFERENCES tc_calendars(id) ON DELETE CASCADE,
 stop_id uuid REFERENCES tc_stops(id) ON DELETE CASCADE, user_id uuid REFERENCES users(id) ON DELETE CASCADE,
 recipient_id uuid, created_at timestamptz NOT NULL DEFAULT now(),
 CHECK ((calendar_id IS NULL) <> (stop_id IS NULL)), CHECK ((user_id IS NULL) <> (recipient_id IS NULL))
);
CREATE UNIQUE INDEX tc_calendar_user_grant ON tc_grants(calendar_id,user_id) WHERE calendar_id IS NOT NULL AND user_id IS NOT NULL;
CREATE UNIQUE INDEX tc_calendar_recipient_grant ON tc_grants(calendar_id,recipient_id) WHERE calendar_id IS NOT NULL AND recipient_id IS NOT NULL;
CREATE UNIQUE INDEX tc_stop_user_grant ON tc_grants(stop_id,user_id) WHERE stop_id IS NOT NULL AND user_id IS NOT NULL;
CREATE UNIQUE INDEX tc_stop_recipient_grant ON tc_grants(stop_id,recipient_id) WHERE stop_id IS NOT NULL AND recipient_id IS NOT NULL;
CREATE TABLE tc_collaborators (
 id uuid PRIMARY KEY, calendar_id uuid REFERENCES tc_calendars(id) ON DELETE CASCADE,
 stop_id uuid REFERENCES tc_stops(id) ON DELETE CASCADE, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 can_manage_permissions boolean NOT NULL DEFAULT false, accepted boolean NOT NULL DEFAULT false,
 created_at timestamptz NOT NULL DEFAULT now(), CHECK ((calendar_id IS NULL) <> (stop_id IS NULL))
);
CREATE UNIQUE INDEX tc_calendar_collaborator ON tc_collaborators(calendar_id,user_id) WHERE calendar_id IS NOT NULL;
CREATE UNIQUE INDEX tc_stop_collaborator ON tc_collaborators(stop_id,user_id) WHERE stop_id IS NOT NULL;
CREATE INDEX tc_collaborator_user_idx ON tc_collaborators(user_id);
CREATE TABLE tc_participations (
 id uuid PRIMARY KEY, stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 status text NOT NULL CHECK (status IN ('interested','requested','joined','needs_reconfirmation','declined','withdrawn','removed','cancelled')),
 start_date text NOT NULL DEFAULT '', end_date text NOT NULL DEFAULT '',
 calendar_link_token_hash text NOT NULL DEFAULT '', stop_link_token_hash text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(stop_id,user_id)
);
CREATE INDEX tc_participation_user_idx ON tc_participations(user_id);
