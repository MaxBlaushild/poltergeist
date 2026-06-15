import { useEffect, useState } from 'react';
import { gmSetQuizOpen, gmListQuizSubmissions } from '../../gmApi';
import type { GMQuizSubmission } from '../../gmApi';
import type { GameState } from '../../types';
import { Card } from './GameSection';

export const QuizSection = ({
  state,
  onChange,
}: {
  state: GameState | null;
  onChange: () => void;
}) => {
  const [subs, setSubs] = useState<GMQuizSubmission[]>([]);
  const [busy, setBusy] = useState(false);

  const loadSubs = () =>
    gmListQuizSubmissions()
      .then((d) => setSubs(d.submissions || []))
      .catch(() => {});
  useEffect(() => {
    loadSubs();
    const id = setInterval(loadSubs, 6000);
    return () => clearInterval(id);
  }, []);

  const toggle = async () => {
    if (!state) return;
    setBusy(true);
    try {
      await gmSetQuizOpen(!state.quizOpen);
      onChange();
    } finally {
      setBusy(false);
    }
  };

  // Group submissions by question for readable review.
  const byQuestion = new Map<number, { prompt: string; rows: GMQuizSubmission[] }>();
  for (const s of subs) {
    if (!byQuestion.has(s.ordinal)) byQuestion.set(s.ordinal, { prompt: s.prompt, rows: [] });
    byQuestion.get(s.ordinal)!.rows.push(s);
  }
  const groups = [...byQuestion.entries()].sort((a, b) => a[0] - b[0]);

  return (
    <div className="flex flex-col gap-6">
      <Card title="End Quiz">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-bone">
              The quiz is{' '}
              <span className={state?.quizOpen ? 'text-green-400' : 'text-blood-bright'}>
                {state?.quizOpen ? 'OPEN' : 'CLOSED'}
              </span>
            </p>
            <p className="text-bone/50 text-sm">
              Opening it takes over every player's screen. Correct answers auto-apply House Favor.
            </p>
          </div>
          <button
            onClick={toggle}
            disabled={busy || !state}
            className={`px-5 py-3 rounded-md uppercase tracking-[0.15em] text-sm disabled:opacity-40 ${
              state?.quizOpen ? 'border border-blood/50 text-bone/70' : 'bg-blood text-bone'
            }`}
          >
            {state?.quizOpen ? 'Close' : 'Open'}
          </button>
        </div>
      </Card>

      {groups.length === 0 ? (
        <p className="text-bone/50">No answers submitted yet.</p>
      ) : (
        groups.map(([ordinal, g]) => (
          <Card key={ordinal} title={`Q${ordinal}. ${g.prompt}`}>
            <div className="flex flex-col gap-2">
              {g.rows.map((r) => (
                <div key={r.id} className="flex items-start gap-3 text-sm">
                  <span className="text-bone/60 w-40 shrink-0">
                    {r.characterName}
                    {r.guestLabel ? ` (${r.guestLabel})` : ''}
                  </span>
                  <span className="text-bone flex-1">{r.answer || '—'}</span>
                  {r.isCorrect === true && <span className="text-green-400 text-xs">✓</span>}
                  {r.isCorrect === false && <span className="text-blood-bright text-xs">✗</span>}
                </div>
              ))}
            </div>
          </Card>
        ))
      )}
    </div>
  );
};
