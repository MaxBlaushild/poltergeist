import { useEffect, useRef, useState } from 'react';
import { getQuiz, submitQuizPart1, submitQuizPart2Answer, getToken } from '../api';
import type { QuizResponse, QuizPart2 } from '../types';
import { VampireMark } from './VampireMark';

const PART1_SECONDS = 5 * 60; // ~5-minute window
const P1_DRAFT = 'vampireQuizP1Draft';

// Full-screen end-quiz takeover for the given part. Part 1 is a timed open-end
// response; Part 2 is silent multiple choice. Answers lock on submit.
export const QuizTakeover = ({ part, onDone }: { part: 1 | 2; onDone: () => void }) => {
  const token = getToken() || '';
  const [data, setData] = useState<QuizResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getQuiz(token)
      .then(setData)
      .catch(() => setError('The quiz could not be loaded.'));
  }, [token]);

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-blood-ink">
      <div className="min-h-full flex items-center justify-center px-4 py-10">
        <div className="w-full max-w-lg text-center">
          <VampireMark className="w-12 h-12 mx-auto mb-4" />
          <p className="text-xs uppercase tracking-[0.4em] text-gold mb-2">The Reckoning</p>
          {error && <p className="text-blood-bright">{error}</p>}
          {!data && !error && <p className="text-bone/70">Summoning the questions…</p>}
          {data && part === 1 && <Part1 data={data} token={token} onDone={onDone} />}
          {data && part === 2 && <Part2 data={data} token={token} onDone={onDone} />}
        </div>
      </div>
    </div>
  );
};

const Part1 = ({
  data,
  token,
  onDone,
}: {
  data: QuizResponse;
  token: string;
  onDone: () => void;
}) => {
  const [answer, setAnswer] = useState(
    () => localStorage.getItem(P1_DRAFT) ?? data.part1.answer ?? ''
  );
  const [done, setDone] = useState(data.part1.submitted);
  const [busy, setBusy] = useState(false);
  const submittedRef = useRef(done);

  // Countdown from openedAt + 5 minutes.
  const deadline = data.part1.openedAt
    ? new Date(data.part1.openedAt).getTime() + PART1_SECONDS * 1000
    : null;
  const [remaining, setRemaining] = useState(
    deadline ? Math.max(0, Math.round((deadline - Date.now()) / 1000)) : PART1_SECONDS
  );

  useEffect(() => {
    localStorage.setItem(P1_DRAFT, answer);
  }, [answer]);

  const submit = async () => {
    if (submittedRef.current || busy) return;
    submittedRef.current = true;
    setBusy(true);
    try {
      await submitQuizPart1(token, answer);
      localStorage.removeItem(P1_DRAFT);
      setDone(true);
    } catch {
      submittedRef.current = false;
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => {
    if (done || !deadline) return;
    const id = setInterval(() => {
      const rem = Math.max(0, Math.round((deadline - Date.now()) / 1000));
      setRemaining(rem);
      if (rem <= 0) {
        clearInterval(id);
        submit(); // auto-submit when time runs out
      }
    }, 1000);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [done, deadline]);

  if (done) return <Sealed onDone={onDone} />;

  const mm = String(Math.floor(remaining / 60)).padStart(2, '0');
  const ss = String(remaining % 60).padStart(2, '0');

  return (
    <>
      <h1 className="font-display text-3xl font-bold text-bone mb-1">Part One</h1>
      <p className={`text-2xl font-semibold mb-4 ${remaining <= 30 ? 'text-blood-bright' : 'text-bone/80'}`}>
        {mm}:{ss}
      </p>
      <p className="text-bone/85 mb-5 text-left leading-relaxed">{data.part1.prompt}</p>
      <textarea
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
        rows={8}
        placeholder="Your testimony…"
        className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
      />
      <button
        onClick={submit}
        disabled={busy}
        className="mt-5 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
      >
        {busy ? 'Sealing…' : 'Seal my testimony'}
      </button>
    </>
  );
};

// Part 2 is answered one question at a time. The server sends only the current
// question and locks each answer on submit — you can't peek ahead or go back.
const Part2 = ({
  data,
  token,
  onDone,
}: {
  data: QuizResponse;
  token: string;
  onDone: () => void;
}) => {
  const [part2, setPart2] = useState<QuizPart2>(data.part2);
  const [selected, setSelected] = useState('');
  const [busy, setBusy] = useState(false);

  const q = part2.questions[0]; // the current question, or undefined when done
  const total = part2.total ?? part2.questions.length;
  const answered = part2.answered ?? 0;
  const done = part2.submitted || !q;
  const isLast = answered + 1 >= total;
  const ready = q ? (q.type === 'number' ? selected !== '' : selected !== '') : false;

  const submit = async () => {
    if (!q || busy || !ready) return;
    setBusy(true);
    try {
      await submitQuizPart2Answer(token, q.id, selected);
      setSelected('');
      const fresh = await getQuiz(token);
      setPart2(fresh.part2);
    } finally {
      setBusy(false);
    }
  };

  if (done) return <Sealed onDone={onDone} />;

  return (
    <>
      <h1 className="font-display text-3xl font-bold text-bone mb-1">Part Two</h1>
      <p className="text-bone/70 mb-6">
        Question {answered + 1} of {total} · each answer locks as you go — no going back.
      </p>
      <div className="rounded-lg border border-blood/40 bg-black/40 p-5 text-left">
        <p className="text-bone text-lg mb-4">{q.prompt}</p>
        {q.type === 'number' ? (
          <input
            type="number"
            inputMode="numeric"
            min={0}
            value={selected}
            onChange={(e) => setSelected(e.target.value.replace(/[^0-9]/g, ''))}
            placeholder="Number of Blood Tokens"
            className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone text-lg"
          />
        ) : (
          <div className="flex flex-col gap-2">
            {q.options.map((opt) => (
              <label
                key={opt}
                className={`flex items-center gap-3 rounded-md border p-3 cursor-pointer transition-colors ${
                  selected === opt
                    ? 'border-blood-bright bg-blood/20 text-bone'
                    : 'border-blood/30 text-bone/80 hover:text-bone'
                }`}
              >
                <input
                  type="radio"
                  name={q.id}
                  checked={selected === opt}
                  onChange={() => setSelected(opt)}
                />
                {opt}
              </label>
            ))}
          </div>
        )}
      </div>
      <button
        onClick={submit}
        disabled={busy || !ready}
        className="mt-6 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
      >
        {busy ? 'Sealing…' : isLast ? 'Lock final answer' : 'Lock & continue'}
      </button>
    </>
  );
};

const Sealed = ({ onDone }: { onDone: () => void }) => (
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
);
