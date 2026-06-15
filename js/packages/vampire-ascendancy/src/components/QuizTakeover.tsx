import { useEffect, useState } from 'react';
import { getQuiz, submitQuiz, getToken } from '../api';
import type { QuizQuestion } from '../types';
import { VampireMark } from './VampireMark';

// Full-screen end-quiz takeover. While the quiz is open and unanswered it can't
// be dismissed; once submitted (locked), the player may return to the app.
const DRAFT_KEY = 'vampireQuizDraft';

export const QuizTakeover = ({ onDone }: { onDone: () => void }) => {
  const token = getToken() || '';
  const [questions, setQuestions] = useState<QuizQuestion[] | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getQuiz(token)
      .then((d) => {
        setQuestions(d.questions);
        // Restore any unsent draft so a reload/blip mid-quiz keeps typed answers.
        let saved: Record<string, string> = {};
        try {
          saved = JSON.parse(localStorage.getItem(DRAFT_KEY) || '{}');
        } catch {
          /* ignore */
        }
        const init: Record<string, string> = {};
        d.questions.forEach((q) => (init[q.id] = saved[q.id] ?? q.answer ?? ''));
        setAnswers(init);
        if (d.submitted) setDone(true);
      })
      .catch(() => setError('The quiz could not be loaded.'));
  }, [token]);

  // Persist drafts as the player types.
  useEffect(() => {
    if (!done && Object.keys(answers).length) {
      localStorage.setItem(DRAFT_KEY, JSON.stringify(answers));
    }
  }, [answers, done]);

  const submit = async () => {
    if (busy) return;
    setBusy(true);
    setError(null);
    try {
      await submitQuiz(
        token,
        Object.entries(answers).map(([questionId, answer]) => ({ questionId, answer }))
      );
      localStorage.removeItem(DRAFT_KEY);
      setDone(true);
    } catch {
      setError('Your answers were not accepted. Try again.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-blood-ink">
      <div className="min-h-full flex items-center justify-center px-4 py-10">
        <div className="w-full max-w-lg text-center">
          <VampireMark className="w-12 h-12 mx-auto mb-4" />
          <p className="text-xs uppercase tracking-[0.4em] text-gold mb-2">The Reckoning</p>

          {done ? (
            <>
              <h1 className="font-display text-3xl font-bold text-bone mb-3">Your answers are sealed</h1>
              <p className="text-bone/80 mb-8">The court has recorded your testimony.</p>
              <button
                onClick={onDone}
                className="px-8 py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright"
              >
                Return
              </button>
            </>
          ) : !questions ? (
            <p className="text-bone/70">Summoning the questions…</p>
          ) : (
            <>
              <h1 className="font-display text-3xl font-bold text-bone mb-1">The End Quiz</h1>
              <p className="text-bone/70 mb-6">Answer all you can — your testimony locks once sent.</p>

              <div className="flex flex-col gap-5 text-left">
                {questions.map((q, i) => (
                  <div key={q.id} className="rounded-lg border border-blood/40 bg-black/40 p-5">
                    <p className="text-bone mb-3">
                      <span className="text-gold mr-2">{i + 1}.</span>
                      {q.prompt}
                    </p>
                    {q.questionType === 'multiple_choice' ? (
                      <div className="flex flex-col gap-2">
                        {q.options.map((opt) => (
                          <label
                            key={opt}
                            className={`flex items-center gap-3 rounded-md border p-3 cursor-pointer transition-colors ${
                              answers[q.id] === opt
                                ? 'border-blood-bright bg-blood/20 text-bone'
                                : 'border-blood/30 text-bone/80 hover:text-bone'
                            }`}
                          >
                            <input
                              type="radio"
                              name={q.id}
                              checked={answers[q.id] === opt}
                              onChange={() => setAnswers((a) => ({ ...a, [q.id]: opt }))}
                            />
                            {opt}
                          </label>
                        ))}
                      </div>
                    ) : (
                      <textarea
                        value={answers[q.id] || ''}
                        onChange={(e) => setAnswers((a) => ({ ...a, [q.id]: e.target.value }))}
                        rows={3}
                        placeholder="Your testimony…"
                        className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
                      />
                    )}
                  </div>
                ))}
              </div>

              {error && <p className="text-blood-bright text-sm mt-4">{error}</p>}

              <button
                onClick={submit}
                disabled={busy}
                className="mt-6 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
              >
                {busy ? 'Sealing…' : 'Seal my answers'}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
