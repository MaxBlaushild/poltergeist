import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  listMyToasts,
  hostToast,
  listActiveMysteries,
  listActiveSubplots,
  type MyToast,
  type ActiveMystery,
} from '../platformApi';
import { adminListHouses } from '../superAdminApi';
import { ApiError } from '../api';
import { useUserAuth } from '../userAuth';
import { VampireMark } from './VampireMark';
import { accentFor, houseLabel } from '../theme';

// GET /toasts — every Toast the signed-in user Hosts, Co-Hosts, or plays a
// character in, reached from the landing page once signed in. Wrapped in
// <RequireUser> by App.tsx. An admin card links into the GM console; a
// player card previews the character and links into the player app.
export const MyToasts = () => {
  const { auth, logout } = useUserAuth();
  const navigate = useNavigate();
  const [toasts, setToasts] = useState<MyToast[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  // Checked "by consequence" (same pattern SuperAdmin.tsx's own gate uses) —
  // there's no dedicated "am I a super user" field on the account, so a
  // super-user-only call either succeeds or 403s.
  const [isSuperUser, setIsSuperUser] = useState(false);

  useEffect(() => {
    listMyToasts()
      .then((d) => setToasts(d.instances))
      .catch(() => setError('Could not load your Toasts.'));
    adminListHouses()
      .then(() => setIsSuperUser(true))
      .catch(() => setIsSuperUser(false));
  }, []);

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 py-10">
      <header className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <VampireMark className="w-8 h-8" />
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-gold">My Toasts</p>
            <p className="text-bone/50 text-xs">Signed in as {auth?.user.name}</p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          {isSuperUser && (
            <Link
              to="/admin"
              className="text-gold/80 hover:text-gold text-xs uppercase tracking-[0.2em]"
            >
              Super Admin
            </Link>
          )}
          <button
            onClick={() => {
              logout();
              navigate('/');
            }}
            className="text-bone/50 hover:text-bone text-xs uppercase tracking-[0.2em]"
          >
            Sign out
          </button>
        </div>
      </header>

      <Link
        to="/toasts/new"
        className="block text-center mb-6 px-6 py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright"
      >
        Host another Toast
      </Link>

      {error && <p className="text-blood-bright text-sm">{error}</p>}
      {!toasts && !error && <p className="text-bone/50 text-center">Gathering your Toasts…</p>}
      {toasts && toasts.length === 0 && (
        <p className="text-bone/50 text-center text-sm">
          No Toasts yet — "Host another Toast" above to start your first one, or ask whoever's running one
          to invite you.
        </p>
      )}

      <div className="space-y-3">
        {toasts?.map((t) => (
          <ToastCard key={t.id} toast={t} />
        ))}
      </div>
    </div>
  );
};

const ToastCard = ({ toast }: { toast: MyToast }) => {
  const href = toast.kind === 'admin' ? `/e/${toast.id}/gm` : `/e/${toast.id}`;
  const accent = toast.kind === 'player' ? accentFor(toast.character?.house) : undefined;

  return (
    <Link
      to={href}
      className="block rounded-lg border border-blood/40 bg-black/40 p-4 hover:border-blood transition-colors"
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          {toast.kind === 'player' && toast.character && (
            <TeaserPortrait imageUrl={toast.character.imageUrl} name={toast.character.name} accent={accent} />
          )}
          <div className="min-w-0">
            <p className="font-heading text-bone font-semibold truncate">{toast.name}</p>
            {toast.kind === 'admin' ? (
              <p className="text-bone/50 text-xs mt-1">
                <span className="inline-block px-2 py-0.5 rounded-full border border-gold/50 text-gold uppercase tracking-[0.15em] text-[10px] mr-2">
                  {toast.role === 'owner' ? 'Host' : 'Co-Host'}
                </span>
                {toast.status}
              </p>
            ) : toast.character ? (
              <p className="text-xs mt-1">
                <span className="text-bone/70">You are </span>
                <span className="font-semibold" style={{ color: accent }}>
                  {toast.character.name}
                </span>
                {toast.character.house && (
                  <span className="text-bone/40"> · {houseLabel(toast.character.house)}</span>
                )}
              </p>
            ) : (
              <p className="text-bone/50 text-xs mt-1">Awaiting your seat</p>
            )}
          </div>
        </div>
        <span className="shrink-0 text-gold text-xs uppercase tracking-[0.2em]">Open →</span>
      </div>
    </Link>
  );
};

// Small round character portrait for the player-card sneak peek. Falls back
// to a house-tinted initial when there's no portrait yet.
const TeaserPortrait = ({
  imageUrl,
  name,
  accent,
}: {
  imageUrl?: string;
  name: string;
  accent?: string;
}) => {
  const [failed, setFailed] = useState(false);
  if (imageUrl && !failed) {
    return (
      <img
        src={imageUrl}
        alt=""
        onError={() => setFailed(true)}
        className="w-10 h-10 rounded-full object-cover border shrink-0"
        style={{ borderColor: accent }}
      />
    );
  }
  return (
    <div
      className="w-10 h-10 rounded-full flex items-center justify-center border shrink-0 font-display text-sm"
      style={{ borderColor: accent, color: accent }}
    >
      {name.charAt(0)}
    </div>
  );
};

// GET/POST /toasts/new — "Host a Toast": just a name. Seeds the shared
// library (everything included) and drops the new Host into the GM console.
export const CreateToast = () => {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [mysteryId, setMysteryId] = useState('');
  const [mysteries, setMysteries] = useState<ActiveMystery[] | null>(null);
  const [subplotIds, setSubplotIds] = useState<Set<string>>(new Set());
  const [subplots, setSubplots] = useState<ActiveMystery[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    listActiveMysteries()
      .then((d) => setMysteries(d.mysteries))
      .catch(() => setMysteries([]));
    listActiveSubplots()
      .then((d) => setSubplots(d.subplots))
      .catch(() => setSubplots([]));
  }, []);

  const toggleSubplot = (id: string) =>
    setSubplotIds((cur) => {
      const next = new Set(cur);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const ready = name.trim() && mysteryId;

  const submit = async () => {
    if (!ready || busy) return;
    setBusy(true);
    setError(null);
    try {
      const inst = await hostToast(name.trim(), mysteryId, Array.from(subplotIds));
      navigate(`/e/${inst.id}/gm`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not host this Toast. Try again.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-10">
      <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6">
        <VampireMark className="w-12 h-12 mx-auto mb-3" />
        <h1 className="font-display text-2xl font-bold text-bone text-center mb-1">Host a Toast</h1>
        <p className="text-bone/50 text-sm text-center mb-6">
          Starts fully stocked from the shared story — trim the roster once you know your guest count.
        </p>
        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">What's it called?</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Sarah's 30th Birthday Toast"
          className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone mb-4"
        />

        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Which mystery?</label>
        <p className="text-bone/40 text-xs mb-2">
          Picked once — this can't be changed after the Toast is created.
        </p>
        {mysteries === null ? (
          <p className="text-bone/50 text-sm mb-4">Loading mysteries…</p>
        ) : mysteries.length === 0 ? (
          <p className="text-gold/80 text-sm mb-4">No mysteries available yet — ask a super user to create one.</p>
        ) : (
          <div className="flex flex-col gap-1.5 mb-4 max-h-64 overflow-y-auto">
            {mysteries.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => setMysteryId(m.id)}
                className={`text-left rounded-md border p-3 transition-colors ${
                  mysteryId === m.id
                    ? 'border-blood-bright bg-blood/20'
                    : 'border-blood/30 bg-black/30 hover:border-blood/50'
                }`}
              >
                <p className="text-bone text-sm font-semibold">{m.name}</p>
                {m.summary && <p className="text-bone/60 text-xs mt-0.5">{m.summary}</p>}
              </button>
            ))}
          </div>
        )}

        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">
          Any sub-plots? <span className="normal-case text-bone/40">(optional)</span>
        </label>
        <p className="text-bone/40 text-xs mb-2">
          Pick as many as you like, in addition to the mystery above — equally permanent once
          hosted.
        </p>
        {subplots === null ? (
          <p className="text-bone/50 text-sm mb-4">Loading sub-plots…</p>
        ) : subplots.length === 0 ? (
          <p className="text-bone/40 text-xs mb-4">No sub-plots available yet.</p>
        ) : (
          <div className="flex flex-col gap-1.5 mb-4 max-h-48 overflow-y-auto">
            {subplots.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => toggleSubplot(s.id)}
                className={`text-left rounded-md border p-3 transition-colors ${
                  subplotIds.has(s.id)
                    ? 'border-blood-bright bg-blood/20'
                    : 'border-blood/30 bg-black/30 hover:border-blood/50'
                }`}
              >
                <p className="text-bone text-sm font-semibold">
                  {subplotIds.has(s.id) && <span className="text-gold mr-1.5">✓</span>}
                  {s.name}
                </p>
                {s.summary && <p className="text-bone/60 text-xs mt-0.5">{s.summary}</p>}
              </button>
            ))}
          </div>
        )}

        {error && <p className="text-blood-bright text-sm mb-3">{error}</p>}
        <button
          onClick={submit}
          disabled={busy || !ready}
          className="w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
        >
          {busy ? 'Hosting…' : 'Host this Toast'}
        </button>
      </div>
    </div>
  );
};
