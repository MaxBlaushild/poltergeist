import { useEffect, useState } from 'react';
import {
  gmListAdmins,
  gmInviteAdmin,
  gmDeleteAdminInvite,
  gmRemoveAdmin,
  gmTransferOwnership,
} from '../../gmApi';
import type { GMAdminRow, GMPendingInvite } from '../../gmApi';
import { ApiError } from '../../api';
import { useUserAuth } from '../../userAuth';
import { Card } from './GameSection';

// Co-Hosts tab: invite/remove administrators and hand off hosting. The Host
// (owner) can't be removed directly — only via "Make Host" on someone else,
// which demotes the current Host to Co-Host first (see
// MULTI_TENANT_REQUIREMENTS.md's "Admin management rights").
export const CoHostsSection = () => {
  const { auth } = useUserAuth();
  const [admins, setAdmins] = useState<GMAdminRow[]>([]);
  const [invites, setInvites] = useState<GMPendingInvite[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviting, setInviting] = useState(false);
  const [inviteLink, setInviteLink] = useState<string | null>(null);

  const load = () => {
    gmListAdmins()
      .then((d) => {
        setAdmins(d.admins);
        setInvites(d.pendingInvites);
      })
      .catch(() => setError('Could not load the Co-Host list.'))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const sendInvite = async () => {
    if (!inviteEmail.trim()) return;
    setInviting(true);
    setError(null);
    setInviteLink(null);
    try {
      const res = await gmInviteAdmin(inviteEmail.trim());
      setInviteLink(`${window.location.origin}/accept-invite/${res.token}`);
      setInviteEmail('');
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not send invite.');
    } finally {
      setInviting(false);
    }
  };

  const remove = async (userId: string) => {
    setError(null);
    try {
      await gmRemoveAdmin(userId);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not remove.');
    }
  };

  const makeHost = async (userId: string) => {
    setError(null);
    try {
      await gmTransferOwnership(userId);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not hand off hosting.');
    }
  };

  const revoke = async (id: string) => {
    setError(null);
    try {
      await gmDeleteAdminInvite(id);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not revoke invite.');
    }
  };

  if (loading) return <Card title="Co-Hosts">Loading…</Card>;

  return (
    <div className="flex flex-col gap-4">
      <Card title="Invite a Co-Host">
        <div className="flex gap-2 flex-wrap">
          <input
            type="email"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && sendInvite()}
            placeholder="cohost@example.com"
            className="flex-1 min-w-[200px] rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          />
          <button
            onClick={sendInvite}
            disabled={inviting || !inviteEmail.trim()}
            className="py-2 px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            {inviting ? 'Sending…' : 'Invite'}
          </button>
        </div>
        {inviteLink && (
          <p className="mt-3 text-sm text-bone/70">
            Share this link with them —{' '}
            <button
              onClick={() => navigator.clipboard?.writeText(inviteLink)}
              className="text-gold underline underline-offset-2"
            >
              copy link
            </button>
          </p>
        )}
        {error && <p className="mt-2 text-blood-bright text-sm">{error}</p>}
      </Card>

      <Card title="Hosts & Co-Hosts">
        <div className="flex flex-col gap-2">
          {admins.map((a) => (
            <div
              key={a.userId}
              className="flex items-center justify-between gap-3 rounded-md border border-blood/30 bg-black/30 px-3 py-2"
            >
              <div className="min-w-0">
                <p className="text-bone truncate">
                  {a.name || a.email || a.userId}
                  {a.userId === auth?.user.id && <span className="text-bone/40 ml-1">(you)</span>}
                </p>
                <p className="text-xs text-gold uppercase tracking-[0.15em]">
                  {a.role === 'owner' ? 'Host' : 'Co-Host'}
                </p>
              </div>
              {a.role !== 'owner' && (
                <div className="flex gap-2 shrink-0">
                  <button
                    onClick={() => makeHost(a.userId)}
                    className="text-xs text-gold uppercase tracking-[0.15em]"
                  >
                    Make Host
                  </button>
                  <button
                    onClick={() => remove(a.userId)}
                    className="text-xs text-blood-bright uppercase tracking-[0.15em]"
                  >
                    Remove
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </Card>

      {invites.length > 0 && (
        <Card title="Pending invites">
          <div className="flex flex-col gap-2">
            {invites.map((i) => (
              <div
                key={i.id}
                className="flex items-center justify-between gap-3 rounded-md border border-blood/30 bg-black/30 px-3 py-2"
              >
                <p className="text-bone/70 truncate">{i.email}</p>
                <button
                  onClick={() => revoke(i.id)}
                  className="shrink-0 text-xs text-blood-bright uppercase tracking-[0.15em]"
                >
                  Revoke
                </button>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
};
