import { useEffect, useState } from 'react';
import { adminListSuperUsers, adminAddSuperUser, adminRemoveSuperUser } from '../../superAdminApi';
import type { AdminSuperUser } from '../../superAdminApi';
import { ApiError } from '../../api';
import { useUserAuth } from '../../userAuth';
import { Card } from '../gm/GameSection';

// Grant/revoke edit access to the shared content library. Any current super
// user can grant another — there's no separate "super-super-user" tier.
export const SuperAdminUsers = () => {
  const { auth } = useUserAuth();
  const [users, setUsers] = useState<AdminSuperUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [email, setEmail] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    adminListSuperUsers()
      .then((d) => setUsers(d.superUsers))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const grant = async () => {
    if (!email.trim()) return;
    setBusy(true);
    setError(null);
    try {
      await adminAddSuperUser(email.trim());
      setEmail('');
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not grant access.');
    } finally {
      setBusy(false);
    }
  };

  const revoke = async (userId: string) => {
    setError(null);
    try {
      await adminRemoveSuperUser(userId);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not revoke access.');
    }
  };

  if (loading) return <Card title="Super users">Loading…</Card>;

  return (
    <div className="flex flex-col gap-4">
      <Card title="Grant access">
        <p className="text-bone/50 text-sm mb-3">
          The person must already have an account (email/password or Google sign-up) here.
        </p>
        <div className="flex gap-2 flex-wrap">
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && grant()}
            placeholder="teammate@example.com"
            className="flex-1 min-w-[200px] rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          />
          <button
            onClick={grant}
            disabled={busy || !email.trim()}
            className="py-2 px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            {busy ? 'Granting…' : 'Grant'}
          </button>
        </div>
        {error && <p className="mt-2 text-blood-bright text-sm">{error}</p>}
      </Card>

      <Card title="Super users">
        <div className="flex flex-col gap-2">
          {users.map((u) => (
            <div key={u.userId} className="flex items-center justify-between gap-3 rounded-md border border-blood/30 bg-black/30 px-3 py-2">
              <p className="text-bone truncate">
                {u.name || u.email || u.userId}
                {u.userId === auth?.user.id && <span className="text-bone/40 ml-1">(you)</span>}
              </p>
              <button onClick={() => revoke(u.userId)} className="shrink-0 text-xs text-blood-bright uppercase tracking-[0.15em]">
                Revoke
              </button>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
};
