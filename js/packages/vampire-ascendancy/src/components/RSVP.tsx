import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getPlayerInvite, declinePlayerInvite, acceptPlayerInvite } from '../platformApi';
import type { PlayerInvite } from '../platformApi';
import { ApiError } from '../api';
import { useUserAuth } from '../userAuth';
import { SignInForm } from './SignInForm';
import { accentFor, houseLabel } from '../theme';
import { VampireMark } from './VampireMark';

// A player's invite, reached from the SMS link. Shows the character teaser
// (name, title, house, pre-event bio) and lets them Accept (sign in/up,
// same real-account flow as Hosts) or Decline (no account needed).
export const RSVP = () => {
  const { token } = useParams();
  const { auth } = useUserAuth();
  const navigate = useNavigate();
  const [invite, setInvite] = useState<PlayerInvite | null>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'notfound'>('loading');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [declined, setDeclined] = useState(false);

  useEffect(() => {
    if (!token) {
      setStatus('notfound');
      return;
    }
    getPlayerInvite(token)
      .then((inv) => {
        setInvite(inv);
        setStatus('ready');
      })
      .catch(() => setStatus('notfound'));
  }, [token]);

  const accept = async () => {
    if (!token) return;
    setBusy(true);
    setError(null);
    try {
      const res = await acceptPlayerInvite(token);
      navigate(`/e/${res.instanceId}`);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Could not accept this invite.');
    } finally {
      setBusy(false);
    }
  };

  const decline = async () => {
    if (!token || busy) return;
    if (!window.confirm("Decline this invite? Your host can invite someone else to this character.")) return;
    setBusy(true);
    setError(null);
    try {
      await declinePlayerInvite(token);
      setDeclined(true);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Could not decline this invite.');
    } finally {
      setBusy(false);
    }
  };

  if (status === 'loading') return <Shell>Approaching the gate…</Shell>;
  if (status === 'notfound' || !invite) {
    return (
      <Shell>
        <VampireMark className="w-14 h-14 mx-auto mb-3" />
        <h1 className="font-display text-2xl font-bold text-bone mb-2">Unknown invitation</h1>
        <p className="text-bone/80">This link is not recognized.</p>
      </Shell>
    );
  }
  if (declined || invite.status === 'declined') {
    return (
      <Shell>
        <VampireMark className="w-14 h-14 mx-auto mb-3" />
        <p className="text-bone/80">You've declined this invite.</p>
      </Shell>
    );
  }
  if (invite.status === 'accepted') {
    return (
      <Shell>
        <VampireMark className="w-14 h-14 mx-auto mb-3" />
        <p className="text-bone/80">This invite has already been accepted.</p>
      </Shell>
    );
  }

  const character = invite.character;
  const accent = accentFor(character?.house);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 gap-6">
      <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6 text-center">
        <p className="text-xs uppercase tracking-[0.4em] text-gold mb-4">
          {invite.instanceName || 'The Crimson Toast'}
        </p>
        <p className="text-bone/60 uppercase tracking-[0.3em] text-xs">{invite.guestName}, you are</p>
        <h1 className="mt-1 font-display text-3xl font-bold text-bone">{character?.name}</h1>
        {character?.title && <p className="text-bone/70 italic mt-1">{character.title}</p>}
        {character?.house && (
          <span
            className="inline-block mt-3 px-3 py-1 rounded-full text-xs uppercase tracking-[0.25em] border"
            style={{ color: accent, borderColor: accent }}
          >
            {houseLabel(character.house)}
          </span>
        )}
        {character?.preEventInfo && (
          <p className="mt-4 text-left text-bone/85 leading-relaxed whitespace-pre-wrap">
            {character.preEventInfo}
          </p>
        )}

        {error && <p className="text-blood-bright text-sm mt-4">{error}</p>}

        {auth && (
          <div className="mt-6 flex flex-col gap-3">
            <button
              onClick={accept}
              disabled={busy}
              className="py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
            >
              {busy ? 'Entering…' : 'Accept invite'}
            </button>
            <button
              onClick={decline}
              disabled={busy}
              className="text-bone/50 hover:text-bone uppercase tracking-[0.2em] text-xs disabled:opacity-40"
            >
              Decline
            </button>
          </div>
        )}
      </div>

      {!auth && (
        <div className="w-full max-w-sm flex flex-col items-center gap-4">
          <SignInForm onSignedIn={accept} title="Accept your invite" />
          <button
            onClick={decline}
            disabled={busy}
            className="text-bone/50 hover:text-bone uppercase tracking-[0.2em] text-xs disabled:opacity-40"
          >
            Decline instead
          </button>
        </div>
      )}
    </div>
  );
};

const Shell = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-4">
    <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6 text-center">
      {children}
    </div>
  </div>
);
