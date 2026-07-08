import { useState } from 'react';
import { gmSetUnlock, gmSetAct, gmResetGame, gmExportStandings } from '../../gmApi';
import { ApiError } from '../../api';
import type { GameState } from '../../types';

const ACTS = ['pre_event', 'act1', 'act2', 'act3', 'quiz', 'resolved'];
const ACT_LABEL: Record<string, string> = {
  pre_event: 'Pre-Event',
  act1: 'Act 1',
  act2: 'Act 2',
  act3: 'Act 3',
  quiz: 'Quiz',
  resolved: 'Resolved',
};

export const GameSection = ({
  state,
  onChange,
}: {
  state: GameState | null;
  onChange: () => void;
}) => {
  const [busy, setBusy] = useState(false);

  if (!state) return <p className="text-bone/50">Loading game state…</p>;

  const toggleUnlock = async () => {
    setBusy(true);
    try {
      await gmSetUnlock(!state.contentUnlocked);
      onChange();
    } finally {
      setBusy(false);
    }
  };

  const setAct = async (act: string) => {
    setBusy(true);
    try {
      await gmSetAct(act);
      onChange();
    } finally {
      setBusy(false);
    }
  };

  const exportStandings = async () => {
    setBusy(true);
    try {
      const data = await gmExportStandings();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `vampire-standings-${data.exportedAt.slice(0, 19).replace(/[:T]/g, '-')}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setBusy(false);
    }
  };

  const reset = async (force = false) => {
    if (
      !window.confirm(
        'Reset for a clean playtest?\n\nThis wipes ALL submissions, House Favor, Blood Tokens, quiz answers, and notifications, and re-seals content. Character assignments and player links are kept.\n\n(Scores are archived first and can be recovered.)'
      )
    )
      return;
    setBusy(true);
    try {
      await gmResetGame(force);
      onChange();
    } catch (e) {
      // Live-lock: the server refuses a reset while the game is live. Offer to override.
      if (e instanceof ApiError && e.status === 409) {
        if (window.confirm(`${e.message}\n\nForce the reset anyway?`)) {
          setBusy(false);
          return reset(true);
        }
      } else {
        throw e;
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <Card title="Content Unlock">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-bone">
              Player content is{' '}
              <span className={state.contentUnlocked ? 'text-green-400' : 'text-blood-bright'}>
                {state.contentUnlocked ? 'UNLOCKED' : 'SEALED'}
              </span>
            </p>
            <p className="text-bone/50 text-sm">
              Reveals post-Act-1 context, secrets, and missions to all players.
            </p>
          </div>
          <button
            onClick={toggleUnlock}
            disabled={busy}
            className={`px-5 py-3 rounded-md uppercase tracking-[0.15em] text-sm disabled:opacity-40 ${
              state.contentUnlocked ? 'border border-blood/50 text-bone/70' : 'bg-blood text-bone'
            }`}
          >
            {state.contentUnlocked ? 'Re-seal' : 'Unlock'}
          </button>
        </div>
      </Card>

      <Card title="Act">
        <p className="text-bone/50 text-sm mb-3">
          Current: <span className="text-bone">{ACT_LABEL[state.currentAct]}</span>
        </p>
        <div className="flex flex-wrap gap-2">
          {ACTS.map((a) => (
            <button
              key={a}
              onClick={() => setAct(a)}
              disabled={busy || a === state.currentAct}
              className={`px-4 py-2 rounded-md text-sm ${
                a === state.currentAct
                  ? 'bg-blood text-bone'
                  : 'border border-blood/40 text-bone/70 hover:text-bone'
              } disabled:opacity-60`}
            >
              {ACT_LABEL[a]}
            </button>
          ))}
        </div>
      </Card>

      <Card title="Standings backup">
        <div className="flex items-center justify-between gap-4">
          <p className="text-bone/60 text-sm max-w-sm">
            Download a snapshot of every house's Favor and every player's Blood Tokens. Keep an
            off-system copy so scores are never only in one place.
          </p>
          <button
            onClick={exportStandings}
            disabled={busy}
            className="px-5 py-3 rounded-md border border-gold/60 text-gold uppercase tracking-[0.15em] text-sm hover:bg-gold hover:text-blood-ink disabled:opacity-40 whitespace-nowrap"
          >
            Export
          </button>
        </div>
      </Card>

      <Card title="Playtest">
        <div className="flex items-center justify-between gap-4">
          <p className="text-bone/60 text-sm max-w-sm">
            Reset to a clean game — clears all progress, keeps the roster and player links. Use
            before a fresh run-through. Scores are archived first, and reset is locked once the
            game is live.
          </p>
          <button
            onClick={() => reset()}
            disabled={busy}
            className="px-5 py-3 rounded-md border border-blood text-blood-bright uppercase tracking-[0.15em] text-sm hover:bg-blood hover:text-bone disabled:opacity-40"
          >
            Reset
          </button>
        </div>
      </Card>
    </div>
  );
};

export const Card = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div className="rounded-lg border border-blood/30 bg-black/40 p-5">
    <h2 className="font-heading text-gold text-xs uppercase tracking-[0.3em] mb-3">{title}</h2>
    {children}
  </div>
);
