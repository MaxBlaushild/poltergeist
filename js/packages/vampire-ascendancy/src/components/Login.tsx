import { useEffect, useState } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { getCharacterPublic, listCharacters, login, saveToken, ApiError } from '../api';
import type { PublicCharacter } from '../api';
import { accentFor } from '../theme';
import { VampireMark } from './VampireMark';

// Reached from a /c/<characterId> link: confirm "you are X" then enter the sigil.
export const ConfirmLogin = () => {
  const { characterId } = useParams();
  const [character, setCharacter] = useState<PublicCharacter | null>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'notfound'>('loading');

  useEffect(() => {
    if (!characterId) {
      setStatus('notfound');
      return;
    }
    getCharacterPublic(characterId)
      .then((c) => {
        setCharacter(c);
        setStatus('ready');
      })
      .catch(() => setStatus('notfound'));
  }, [characterId]);

  if (status === 'loading') return <Shell>Approaching the gate…</Shell>;
  if (status === 'notfound' || !character) {
    return (
      <Shell>
        <VampireMark className="w-14 h-14 mx-auto mb-3" />
        <h1 className="font-display text-2xl font-bold text-bone mb-2">Unknown invitation</h1>
        <p className="text-bone/80 mb-4">This link is not recognized.</p>
        <Link to="/login" className="text-gold uppercase tracking-[0.2em] text-sm">
          Select your name →
        </Link>
      </Shell>
    );
  }
  return <LoginForm fixed={character} />;
};

// Reached from the general /login link: pick your name, then enter the sigil.
export const SelectLogin = () => {
  const [characters, setCharacters] = useState<PublicCharacter[] | null>(null);
  useEffect(() => {
    listCharacters().then((d) => setCharacters(d.characters)).catch(() => setCharacters([]));
  }, []);
  if (!characters) return <Shell>Gathering the court…</Shell>;
  return <LoginForm characters={characters} />;
};

const LoginForm = ({
  fixed,
  characters,
}: {
  fixed?: PublicCharacter;
  characters?: PublicCharacter[];
}) => {
  const navigate = useNavigate();
  const [characterId, setCharacterId] = useState(fixed?.id ?? '');
  const [sigil, setSigil] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const current = fixed ?? characters?.find((c) => c.id === characterId);
  const accent = accentFor(current?.house);

  const submit = async () => {
    if (!characterId || !sigil || busy) return;
    setBusy(true);
    setError(null);
    try {
      const { token } = await login(characterId, sigil);
      saveToken(token);
      localStorage.removeItem('vampireTab'); // land on the Summons after logging in
      navigate('/');
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 401
          ? 'That sigil does not match. Try again, or choose a different name.'
          : 'The court could not be reached. Try again.'
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <Shell>
      <VampireMark className="w-12 h-12 mx-auto mb-3" />
      <p className="text-xs uppercase tracking-[0.4em] text-gold mb-4">The Crimson Toast</p>

      {fixed ? (
        <div className="mb-5">
          <p className="text-bone/60 uppercase tracking-[0.3em] text-xs">You are</p>
          <h1 className="font-display text-3xl font-bold text-bone mt-1">{fixed.name}</h1>
          {fixed.house && (
            <span
              className="inline-block mt-2 px-3 py-1 rounded-full text-xs uppercase tracking-[0.25em] border"
              style={{ color: accent, borderColor: accent }}
            >
              House of {fixed.house}
            </span>
          )}
        </div>
      ) : (
        <div className="mb-4 text-left">
          <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">
            Your name
          </label>
          <select
            value={characterId}
            onChange={(e) => {
              setCharacterId(e.target.value);
              setError(null);
            }}
            className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
          >
            <option value="">— Select your name —</option>
            {characters?.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
      )}

      <div className="text-left">
        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Sigil</label>
        <input
          inputMode="numeric"
          value={sigil}
          onChange={(e) => setSigil(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder="Your 4-digit sigil"
          className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone text-center tracking-[0.3em] text-lg placeholder:tracking-normal placeholder:text-sm placeholder:text-bone/30"
        />
      </div>

      {error && <p className="text-blood-bright text-sm mt-3">{error}</p>}

      <button
        onClick={submit}
        disabled={busy || !characterId || !sigil}
        className="mt-5 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
      >
        {busy ? 'Entering…' : 'Enter the court'}
      </button>

      {fixed && (
        <Link
          to="/login"
          className="block mt-4 text-bone/50 hover:text-bone uppercase tracking-[0.2em] text-xs"
        >
          This isn't me →
        </Link>
      )}
    </Shell>
  );
};

const Shell = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-4">
    <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6 text-center">
      {children}
    </div>
  </div>
);
