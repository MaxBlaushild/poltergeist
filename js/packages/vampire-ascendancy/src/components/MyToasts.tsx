import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { listMyToasts, hostToast, type MyToast } from '../platformApi';
import { ApiError } from '../api';
import { useUserAuth } from '../userAuth';
import { VampireMark } from './VampireMark';

// GET /toasts — every Toast the signed-in user Hosts or Co-Hosts, reached
// from the landing page once signed in. Wrapped in <RequireUser> by App.tsx.
export const MyToasts = () => {
  const { auth, logout } = useUserAuth();
  const navigate = useNavigate();
  const [toasts, setToasts] = useState<MyToast[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listMyToasts()
      .then((d) => setToasts(d.instances))
      .catch(() => setError('Could not load your Toasts.'));
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
        <button
          onClick={() => {
            logout();
            navigate('/');
          }}
          className="text-bone/50 hover:text-bone text-xs uppercase tracking-[0.2em]"
        >
          Sign out
        </button>
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
          No Toasts yet — "Host another Toast" above to start your first one.
        </p>
      )}

      <div className="space-y-3">
        {toasts?.map((t) => (
          <Link
            key={t.id}
            to={`/e/${t.id}/gm`}
            className="block rounded-lg border border-blood/40 bg-black/40 p-4 hover:border-blood transition-colors"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="font-heading text-bone font-semibold">{t.name}</p>
                <p className="text-bone/50 text-xs mt-1">
                  {t.role === 'owner' ? 'Host' : 'Co-Host'} · {t.status}
                </p>
              </div>
              <span className="text-gold text-xs uppercase tracking-[0.2em]">Open →</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
};

// GET/POST /toasts/new — "Host a Toast": just a name. Seeds the shared
// library (everything included) and drops the new Host into the GM console.
export const CreateToast = () => {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    if (!name.trim() || busy) return;
    setBusy(true);
    setError(null);
    try {
      const inst = await hostToast(name.trim());
      navigate(`/e/${inst.id}/gm`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not host this Toast. Try again.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
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
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder="Sarah's 30th Birthday Toast"
          className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone mb-4"
        />
        {error && <p className="text-blood-bright text-sm mb-3">{error}</p>}
        <button
          onClick={submit}
          disabled={busy || !name.trim()}
          className="w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
        >
          {busy ? 'Hosting…' : 'Host this Toast'}
        </button>
      </div>
    </div>
  );
};
