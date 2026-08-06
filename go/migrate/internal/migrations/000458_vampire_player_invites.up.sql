-- Player invites replace the old "pick your character from a list + enter a
-- sigil" walk-up login. A Host/Co-Host now invites a specific real person
-- (name + phone number) to a specific character; the invite is delivered by
-- SMS with a link to accept (sign in/up with a real account, same as
-- Hosts) or decline. See MULTI_TENANT_REQUIREMENTS.md and the "Player
-- accounts (future phase)" section — this is that phase, arriving earlier
-- than expected.
CREATE TABLE IF NOT EXISTS vampire_player_invites (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  character_id UUID NOT NULL REFERENCES vampire_characters(id) ON DELETE CASCADE,
  guest_name TEXT NOT NULL DEFAULT '',
  phone_number TEXT NOT NULL DEFAULT '',
  token TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'pending', -- pending | accepted | declined
  invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
  -- Set once accepted; the same user id lands on vampire_players.user_id.
  accepted_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  responded_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS vampire_player_invites_instance_idx ON vampire_player_invites(instance_id);
CREATE INDEX IF NOT EXISTS vampire_player_invites_character_idx ON vampire_player_invites(character_id);
-- At most one open (pending) invite per character per instance at a time —
-- an admin re-inviting someone else must revoke or wait out the existing
-- one first, so two people can't both think they're playing the same role.
CREATE UNIQUE INDEX IF NOT EXISTS vampire_player_invites_one_pending_idx
  ON vampire_player_invites(instance_id, character_id) WHERE status = 'pending';

-- vampire_players: a row is now created when an invite is accepted (not
-- pre-provisioned as an empty slot), and identifies its holder by real
-- account instead of an opaque token.
ALTER TABLE vampire_players ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL;
-- One character per user per instance.
CREATE UNIQUE INDEX IF NOT EXISTS vampire_players_instance_user_idx
  ON vampire_players(instance_id, user_id) WHERE user_id IS NOT NULL;
-- token/password-era sigil auth is retired; token is no longer required for
-- new rows. Column kept (not dropped) for any pre-existing rows and to
-- avoid a disruptive drop — the application no longer reads or writes it.
ALTER TABLE vampire_players ALTER COLUMN token DROP NOT NULL;
COMMENT ON COLUMN vampire_players.token IS 'DEPRECATED — player auth moved to real accounts (user_id). Do not read/write.';

-- vampire_instance_characters: "included" and "sigil" were the walk-up
-- login's mechanisms (which characters are choosable, and their PIN).
-- Character availability is now implicit — a character is "in" an instance
-- once someone has an accepted invite for it — so these are retired the
-- same way vampire_characters.password/image_url were: commented as
-- deprecated rather than dropped, so nothing already-written breaks.
-- image_url (portrait) is unaffected and stays in active use.
COMMENT ON COLUMN vampire_instance_characters.included IS 'DEPRECATED — character availability is now implicit (has an accepted invite or not). Do not read/write.';
COMMENT ON COLUMN vampire_instance_characters.sigil IS 'DEPRECATED — player auth moved to real accounts. Do not read/write.';
