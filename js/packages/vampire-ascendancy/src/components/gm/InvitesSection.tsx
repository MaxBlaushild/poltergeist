import { useEffect, useState } from 'react';
import { gmListInvites, gmCreateInvite, gmDeleteInvite, gmResendInvite, gmGetCharacterPool } from '../../gmApi';
import type { GMInvite } from '../../gmApi';
import { ApiError } from '../../api';
import { Card } from './GameSection';

// Invites tab: the only way a new person joins this Toast as a player. The
// Host/Co-Host names a real person and a text goes out with an RSVP link —
// accept sets up their account and drops them into the app, where they
// choose their own character from the Character Pool tab's curated set
// (decline frees nothing character-specific, since none was reserved).
// See CharacterPoolSection for curating who's choosable, and PlayersSection
// for managing people who've already accepted.
export const InvitesSection = () => {
  const [invites, setInvites] = useState<GMInvite[]>([]);
  const [poolSize, setPoolSize] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    Promise.all([gmListInvites(), gmGetCharacterPool()])
      .then(([i, p]) => {
        setInvites(i.invites);
        setPoolSize(p.characterIds.length);
      })
      .catch(() => setError('Could not load invites.'))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  if (loading) return <Card title="Invites">Loading…</Card>;

  const pending = invites.filter((i) => i.status === 'pending');
  const resolved = invites.filter((i) => i.status !== 'pending');
  // "Invited" counts pending + accepted — still-live invites, not ones
  // that were declined or revoked — against how many characters a player
  // could actually end up choosing.
  const invitedCount = pending.length + invites.filter((i) => i.status === 'accepted').length;

  return (
    <div className="flex flex-col gap-4">
      <NewInviteForm onCreated={load} onError={setError} />

      {poolSize !== null && (
        <p className="text-sm text-bone/70 -mt-1">
          <span className="text-bone font-semibold">{invitedCount}</span> invited ·{' '}
          <span className="text-bone font-semibold">{poolSize}</span> character{poolSize === 1 ? '' : 's'} in the
          pool
          {poolSize === 0 && (
            <span className="text-gold/80"> — visit the Character Pool tab to make some choosable</span>
          )}
          {poolSize > 0 && invitedCount > poolSize && (
            <span className="text-gold/80"> — more people invited than characters available</span>
          )}
        </p>
      )}
      {error && <p className="text-blood-bright text-sm">{error}</p>}

      {pending.length > 0 && (
        <Card title={`Pending (${pending.length})`}>
          <div className="flex flex-col gap-2">
            {pending.map((inv) => (
              <InviteRow key={inv.id} invite={inv} onChanged={load} onError={setError} />
            ))}
          </div>
        </Card>
      )}

      {resolved.length > 0 && (
        <Card title="Accepted & declined">
          <div className="flex flex-col gap-2">
            {resolved.map((inv) => (
              <InviteRow key={inv.id} invite={inv} onChanged={load} onError={setError} />
            ))}
          </div>
        </Card>
      )}

      {invites.length === 0 && (
        <p className="text-bone/50 text-sm">No invites sent yet — add one above to bring in your first player.</p>
      )}
    </div>
  );
};

const NewInviteForm = ({
  onCreated,
  onError,
}: {
  onCreated: () => void;
  onError: (msg: string | null) => void;
}) => {
  const [guestName, setGuestName] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [busy, setBusy] = useState(false);
  const [warning, setWarning] = useState<string | null>(null);

  const ready = guestName.trim() && phoneNumber.trim();

  const send = async () => {
    if (!ready || busy) return;
    setBusy(true);
    onError(null);
    setWarning(null);
    try {
      const res = await gmCreateInvite(guestName.trim(), phoneNumber.trim());
      if (res.warning) setWarning(res.warning);
      setGuestName('');
      setPhoneNumber('');
      onCreated();
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'Could not send invite.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title="Invite a player">
      <div className="flex flex-col gap-2">
        <div className="flex gap-2 flex-wrap">
          <input
            value={guestName}
            onChange={(e) => setGuestName(e.target.value)}
            placeholder="Guest's name"
            className="flex-1 min-w-[160px] rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          />
          <input
            type="tel"
            value={phoneNumber}
            onChange={(e) => setPhoneNumber(e.target.value)}
            placeholder="Phone number, e.g. +15551234567"
            className="flex-1 min-w-[180px] rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          />
        </div>
        <p className="text-[11px] text-bone/40">
          They'll pick their own character after accepting — see the Character Pool tab to control which
          ones are offered.
        </p>
        <button
          onClick={send}
          disabled={!ready || busy}
          className="py-2 px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40 self-start"
        >
          {busy ? 'Sending…' : 'Send invite'}
        </button>
        {warning && <p className="text-gold text-sm">{warning}</p>}
      </div>
    </Card>
  );
};

const InviteRow = ({
  invite,
  onChanged,
  onError,
}: {
  invite: GMInvite;
  onChanged: () => void;
  onError: (msg: string | null) => void;
}) => {
  const [busy, setBusy] = useState<'resend' | 'delete' | null>(null);

  const resend = async () => {
    setBusy('resend');
    onError(null);
    try {
      await gmResendInvite(invite.id);
      onChanged();
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'Could not resend.');
    } finally {
      setBusy(null);
    }
  };

  const remove = async () => {
    setBusy('delete');
    onError(null);
    try {
      await gmDeleteInvite(invite.id);
      onChanged();
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'Could not remove.');
    } finally {
      setBusy(null);
    }
  };

  const statusStyle: Record<string, string> = {
    pending: 'text-gold',
    accepted: 'text-green-400',
    declined: 'text-blood-bright',
  };

  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-blood/30 bg-black/30 px-3 py-2">
      <div className="min-w-0">
        <p className="text-bone truncate">{invite.guestName}</p>
        <p className={`text-xs uppercase tracking-[0.15em] ${statusStyle[invite.status] || 'text-bone/50'}`}>
          {invite.status}
        </p>
      </div>
      <div className="flex gap-2 shrink-0">
        {invite.status === 'pending' && (
          <button
            onClick={resend}
            disabled={busy !== null}
            className="text-xs text-gold uppercase tracking-[0.15em] disabled:opacity-40"
          >
            {busy === 'resend' ? 'Sending…' : 'Resend'}
          </button>
        )}
        <button
          onClick={remove}
          disabled={busy !== null}
          className="text-xs text-blood-bright uppercase tracking-[0.15em] disabled:opacity-40"
        >
          {busy === 'delete' ? 'Removing…' : invite.status === 'pending' ? 'Revoke' : 'Clear'}
        </button>
      </div>
    </div>
  );
};
