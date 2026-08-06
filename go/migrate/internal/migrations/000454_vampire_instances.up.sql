-- Vampire Ascendancy multi-tenancy, part 1: instances themselves and who
-- administers them. See go/vampire-ascendancy/docs/MULTI_TENANT_REQUIREMENTS.md.

-- One instance = one group's run of the event ("hosting a Toast"). Content
-- (characters/items/quiz) stays a shared global library; an instance only
-- selects a subset of it (see 000456) and owns its own play-state (000455).
CREATE TABLE IF NOT EXISTS vampire_instances (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active', -- active | archived
  -- Nullable: the legacy instance created by the 000455 backfill has no
  -- single creator on record. Every instance created through the app from
  -- here on always sets this.
  created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Who can administer an instance ("Host" / "Co-Host" in the UI). Both roles
-- have identical game-management powers; the distinction only matters for
-- the one guardrail below.
CREATE TABLE IF NOT EXISTS vampire_instance_admins (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'admin', -- owner | admin
  UNIQUE (instance_id, user_id)
);
CREATE INDEX IF NOT EXISTS vampire_instance_admins_instance_idx ON vampire_instance_admins(instance_id);
CREATE INDEX IF NOT EXISTS vampire_instance_admins_user_idx ON vampire_instance_admins(user_id);
-- Exactly one owner per instance. The app enforces the "owner can't be
-- removed by someone else, only via transfer" rule; this index just makes
-- "two owners at once" structurally impossible.
CREATE UNIQUE INDEX IF NOT EXISTS vampire_instance_admins_one_owner_idx
  ON vampire_instance_admins(instance_id) WHERE role = 'owner';

-- Invite-by-email flow: the invitee may not have a User account yet, so this
-- is keyed by email + a bearer token rather than a user_id until accepted.
CREATE TABLE IF NOT EXISTS vampire_instance_admin_invites (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS vampire_instance_admin_invites_instance_idx ON vampire_instance_admin_invites(instance_id);
CREATE INDEX IF NOT EXISTS vampire_instance_admin_invites_email_idx ON vampire_instance_admin_invites(email);
