import { useEffect, useState } from 'react';
import { gmListInvites, gmCreateInvite, gmDeleteInvite, gmResendInvite, gmListCharacters, gmListPlayers } from '../../gmApi';
import type { GMInvite, GMCharacter } from '../../gmApi';
import { ApiError } from '../../api';
import { Card } from './GameSection';
import { CharacterBrowser } from './CharacterBrowser';

// Invites tab: the only way a new person joins this Toast as a player. The
// Host/Co-Host names a real person, picks a character for them, and a text
// goes out with their bio + an RSVP link (accept sets up their account,
// decline frees the character for someone else). See PlayersSection for
// managing people who've already accepted.
export const InvitesSection = () => {
  const [invites, setInvites] = useState<GMInvite[]>([]);
  const [characters, setCharacters] = useState<GMCharacter[]>([]);
  const [takenCharacterIds, setTakenCharacterIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    Promise.all([gmListInvites(), gmListCharacters(), gmListPlayers()])
      .then(([i, c, p]) => {
        setInvites(i.invites);
        setCharacters(c.characters);
        const taken = new Set<string>();
        // A character already held by an active player, or already the
        // target of a pending invite, isn't offered again — avoids double-
        // booking the same role.
        p.players.forEach((pl) => {
          if (pl.character && pl.active) taken.add(pl.character.id);
        });
        i.invites.forEach((inv) => {
          if (inv.status === 'pending' && inv.character) taken.add(inv.character.id);
        });
        setTakenCharacterIds(taken);
      })
      .catch(() => setError('Could not load invites.'))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  if (loading) return <Card title="Invites">Loading…</Card>;

  const assignable = characters.filter((c) => c.roleType === 'player' && !takenCharacterIds.has(c.id));

  const pending = invites.filter((i) => i.status === 'pending');
  const resolved = invites.filter((i) => i.status !== 'pending');

  return (
    <div className="flex flex-col gap-4">
      <NewInviteForm characters={assignable} onCreated={load} onError={setError} />
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
  characters,
  onCreated,
  onError,
}: {
  characters: GMCharacter[];
  onCreated: () => void;
  onError: (msg: string | null) => void;
}) => {
  const [guestName, setGuestName] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [characterId, setCharacterId] = useState('');
  const [busy, setBusy] = useState(false);
  const [warning, setWarning] = useState<string | null>(null);

  const ready = guestName.trim() && phoneNumber.trim() && characterId;

  const send = async () => {
    if (!ready || busy) return;
    setBusy(true);
    onError(null);
    setWarning(null);
    try {
      const res = await gmCreateInvite(guestName.trim(), phoneNumber.trim(), characterId);
      if (res.warning) setWarning(res.warning);
      setGuestName('');
      setPhoneNumber('');
      setCharacterId('');
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
        <CharacterBrowser characters={characters} selectedId={characterId} onSelect={(c) => setCharacterId(c.id)} />
        {characters.length === 0 && (
          <p className="text-bone/40 text-xs">
            Every character is already assigned or has a pending invite.
          </p>
        )}
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
        <p className="text-bone truncate">
          {invite.guestName}
          <span className="text-bone/40 mx-2">•</span>
          <span className="text-bone/60">{invite.character?.name ?? 'Unknown character'}</span>
        </p>
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
