import { useEffect, useState } from 'react';
import {
  getGMAuth,
  setGMAuth,
  clearGMAuth,
  gmGetState,
} from '../../gmApi';
import type { GameState } from '../../types';
import { GameSection } from './GameSection';
import { SubmissionsSection } from './SubmissionsSection';
import { AwardsSection } from './AwardsSection';
import { PlayersSection } from './PlayersSection';
import { BroadcastSection } from './BroadcastSection';
import { QuizSection } from './QuizSection';

const GM_NAMES = ['Ali', 'Max', 'Ngozi', 'Jon'];
type Tab = 'game' | 'submissions' | 'awards' | 'broadcast' | 'quiz' | 'players';

export const GMAdmin = () => {
  const [authed, setAuthed] = useState(false);
  const [checking, setChecking] = useState(true);

  // Re-validate any stored passcode on load.
  useEffect(() => {
    const { pass } = getGMAuth();
    if (!pass) {
      setChecking(false);
      return;
    }
    gmGetState()
      .then(() => setAuthed(true))
      .catch(() => clearGMAuth())
      .finally(() => setChecking(false));
  }, []);

  if (checking) return <Centered>Verifying the seal…</Centered>;
  if (!authed) return <GMLogin onAuthed={() => setAuthed(true)} />;
  return <GMConsole onLogout={() => { clearGMAuth(); setAuthed(false); }} />;
};

const GMLogin = ({ onAuthed }: { onAuthed: () => void }) => {
  const [name, setName] = useState(GM_NAMES[0]);
  const [pass, setPass] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setBusy(true);
    setError(null);
    setGMAuth(pass, name);
    try {
      await gmGetState();
      onAuthed();
    } catch {
      clearGMAuth();
      setError('The passcode was rejected.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold text-center">
          The Crimson Toast
        </p>
        <h1 className="mt-3 font-display text-2xl font-bold text-bone text-center mb-6">
          Court Master
        </h1>

        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">You are</label>
        <select
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full mb-4 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        >
          {GM_NAMES.map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>

        <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Passcode</label>
        <input
          type="password"
          value={pass}
          onChange={(e) => setPass(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          className="w-full mb-4 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        />
        {error && <p className="text-blood-bright text-sm mb-3">{error}</p>}
        <button
          onClick={submit}
          disabled={busy || !pass}
          className="w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
        >
          {busy ? 'Entering…' : 'Enter'}
        </button>
      </div>
    </div>
  );
};

const GMConsole = ({ onLogout }: { onLogout: () => void }) => {
  const [tab, setTab] = useState<Tab>('game');
  const [state, setState] = useState<GameState | null>(null);
  const { name } = getGMAuth();

  const refreshState = () => gmGetState().then(setState).catch(() => {});
  useEffect(() => {
    refreshState();
  }, []);

  const tabs: { id: Tab; label: string }[] = [
    { id: 'game', label: 'Game' },
    { id: 'submissions', label: 'Submissions' },
    { id: 'awards', label: 'Awards' },
    { id: 'broadcast', label: 'Broadcast' },
    { id: 'quiz', label: 'Quiz' },
    { id: 'players', label: 'Players' },
  ];

  return (
    <div className="min-h-screen max-w-3xl mx-auto px-4 py-6">
      <header className="flex items-center justify-between mb-5">
        <div>
          <p className="font-heading text-xs uppercase tracking-[0.35em] text-gold">Court Master</p>
          <p className="text-bone/80 text-sm">Acting as {name}</p>
        </div>
        <button onClick={onLogout} className="text-bone/50 text-sm uppercase tracking-[0.2em]">
          Leave
        </button>
      </header>

      <nav className="flex gap-2 mb-6 flex-wrap">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 rounded-md text-sm uppercase tracking-[0.15em] ${
              tab === t.id ? 'bg-blood text-bone' : 'text-bone/60 border border-blood/40'
            }`}
          >
            {t.label}
          </button>
        ))}
      </nav>

      {tab === 'game' && <GameSection state={state} onChange={refreshState} />}
      {tab === 'submissions' && <SubmissionsSection />}
      {tab === 'awards' && <AwardsSection />}
      {tab === 'broadcast' && <BroadcastSection />}
      {tab === 'quiz' && <QuizSection state={state} onChange={refreshState} />}
      {tab === 'players' && <PlayersSection />}
    </div>
  );
};

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center text-bone/80">
    {children}
  </div>
);
