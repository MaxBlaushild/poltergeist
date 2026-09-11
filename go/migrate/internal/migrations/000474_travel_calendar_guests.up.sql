CREATE TABLE tc_guest_recipients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    name text NOT NULL DEFAULT '',
    user_id uuid REFERENCES users(id),
    verified_at timestamptz,
    participation_notices boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_guest_recipients_user ON tc_guest_recipients(user_id);
ALTER TABLE tc_grants ADD CONSTRAINT tc_grants_guest_recipient_fk
    FOREIGN KEY (recipient_id) REFERENCES tc_guest_recipients(id) ON DELETE CASCADE;

CREATE TABLE tc_guest_challenges (
    id uuid PRIMARY KEY,
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    code_hash text NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_guest_challenges_expiry ON tc_guest_challenges(expires_at);

CREATE TABLE tc_guest_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    purpose text NOT NULL CHECK (purpose IN ('view', 'preferences')),
    verified_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_guest_tokens_expiry ON tc_guest_tokens(expires_at);

CREATE TABLE tc_guest_rate_limits (
    key_hash text PRIMARY KEY,
    count integer NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE TABLE tc_calendar_invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    calendar_id uuid REFERENCES tc_calendars(id) ON DELETE CASCADE,
    stop_id uuid REFERENCES tc_stops(id) ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES users(id),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'sent', 'failed')),
    last_error text NOT NULL DEFAULT '',
    sent_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((calendar_id IS NULL) <> (stop_id IS NULL))
);
CREATE UNIQUE INDEX tc_calendar_invitations_calendar_unique ON tc_calendar_invitations(recipient_id, calendar_id) WHERE calendar_id IS NOT NULL;
CREATE UNIQUE INDEX tc_calendar_invitations_stop_unique ON tc_calendar_invitations(recipient_id, stop_id) WHERE stop_id IS NOT NULL;

CREATE TABLE tc_calendar_subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    calendar_id uuid REFERENCES tc_calendars(id) ON DELETE CASCADE,
    stop_id uuid REFERENCES tc_stops(id) ON DELETE CASCADE,
    status text NOT NULL CHECK (status IN ('active', 'paused', 'unsubscribed')),
    calendar_link_token_hash text NOT NULL DEFAULT '',
    stop_link_token_hash text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((calendar_id IS NULL) <> (stop_id IS NULL))
);
CREATE UNIQUE INDEX tc_calendar_subscriptions_calendar_unique ON tc_calendar_subscriptions(recipient_id, calendar_id) WHERE calendar_id IS NOT NULL;
CREATE UNIQUE INDEX tc_calendar_subscriptions_stop_unique ON tc_calendar_subscriptions(recipient_id, stop_id) WHERE stop_id IS NOT NULL;

-- A calendar follow must not silently resume updates for a revoked stop when
-- its audience is later restored. The follower explicitly opts in again.
CREATE TABLE tc_subscription_exclusions (
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (recipient_id, stop_id)
);

CREATE TABLE tc_calendar_broadcasts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id uuid NOT NULL REFERENCES users(id),
    calendar_id uuid REFERENCES tc_calendars(id) ON DELETE CASCADE,
    stop_ids jsonb NOT NULL DEFAULT '[]',
    message text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE tc_calendar_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcast_id uuid REFERENCES tc_calendar_broadcasts(id) ON DELETE CASCADE,
    recipient_id uuid NOT NULL REFERENCES tc_guest_recipients(id) ON DELETE CASCADE,
    stop_id uuid REFERENCES tc_stops(id) ON DELETE CASCADE,
    participation_id uuid REFERENCES tc_participations(id) ON DELETE CASCADE,
    event text NOT NULL DEFAULT 'broadcast',
    dedupe_key text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'skipped')),
    attempts integer NOT NULL DEFAULT 0,
    last_error text NOT NULL DEFAULT '',
    sent_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_calendar_deliveries_broadcast ON tc_calendar_deliveries(broadcast_id);
CREATE INDEX tc_calendar_deliveries_pending ON tc_calendar_deliveries(status, created_at);

CREATE TABLE tc_calendar_notices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stop_id uuid NOT NULL REFERENCES tc_stops(id) ON DELETE CASCADE,
    participation_id uuid NOT NULL REFERENCES tc_participations(id) ON DELETE CASCADE,
    event text NOT NULL,
    dedupe_key text NOT NULL UNIQUE,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tc_calendar_notices_user ON tc_calendar_notices(user_id, created_at);
