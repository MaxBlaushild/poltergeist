-- Super users: the only accounts allowed to edit the shared content library
-- (characters, houses, items, quiz questions) — separate from Host/Co-Host,
-- which is scoped to one instance and only toggles inclusion. See
-- go/vampire-ascendancy/docs/MULTI_TENANT_REQUIREMENTS.md's "Open
-- inconsistency" note.
CREATE TABLE IF NOT EXISTS vampire_super_users (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Nullable: the first super user is granted by the ops-only
  -- cmd/grant-super-user CLI, which has no "granted by" account to record.
  created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Audit trail for the shared-library editor, mirroring vampire_gm_action_log
-- (which is per-instance and can't record actions that touch every instance
-- at once).
CREATE TABLE IF NOT EXISTS vampire_super_user_action_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb
);
