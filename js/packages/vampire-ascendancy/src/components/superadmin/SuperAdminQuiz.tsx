import { useEffect, useState } from 'react';
import { adminGetMysteryQuiz, adminUpdateMysteryQuiz } from '../../superAdminApi';
import type { GMQuizQuestions } from '../../gmApi';
import { Card } from '../gm/GameSection';

const qInput = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

// Editor for one mystery's quiz content: the Part 1 open-end prompt +
// rubric, and the Part 2 multiple-choice questions (including the answer
// key — that's why this is super-user-only, not just editing: it's a
// spoiler for anyone else's Toast running the same mystery). Saving
// replaces the question set for every instance running this mystery and
// clears any existing answers.
export const SuperAdminQuiz = ({ mysteryId }: { mysteryId: string }) => {
  const [q, setQ] = useState<GMQuizQuestions | null>(null);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    setQ(null);
    adminGetMysteryQuiz(mysteryId).then(setQ).catch(() => setNote('Could not load questions.'));
  }, [mysteryId]);

  const updMc = (i: number, patch: Partial<GMQuizQuestions['part2'][number]>) =>
    setQ((prev) => (prev ? { ...prev, part2: prev.part2.map((x, j) => (j === i ? { ...x, ...patch } : x)) } : prev));

  const save = async () => {
    if (!q) return;
    if (
      !window.confirm(
        "Save quiz questions?\n\nThis replaces this mystery's question set for every Toast running it and clears any existing quiz answers. Do this before any quiz runs."
      )
    )
      return;
    setBusy(true);
    setNote(null);
    try {
      await adminUpdateMysteryQuiz(mysteryId, q);
      setNote('Saved.');
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  if (!q) return <Card title="Quiz questions">{note || 'Loading…'}</Card>;

  return (
    <Card title="Quiz questions (this mystery — affects every Toast running it)">
      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <span className="text-[11px] uppercase tracking-[0.15em] text-gold">Part 1 — open-end</span>
          <textarea
            className={qInput}
            rows={2}
            placeholder="Prompt"
            value={q.part1.prompt}
            onChange={(e) => setQ({ ...q, part1: { ...q.part1, prompt: e.target.value } })}
          />
          <textarea
            className={qInput}
            rows={4}
            placeholder="Rubric — the canonical truth / grading guide the AI scores against"
            value={q.part1.rubric}
            onChange={(e) => setQ({ ...q, part1: { ...q.part1, rubric: e.target.value } })}
          />
          <label className="text-xs text-bone/50 flex items-center gap-2">
            Max BT
            <input
              type="number"
              className="w-20 rounded-md bg-black/60 border border-blood/40 p-1.5 text-bone text-center"
              value={q.part1.maxBt}
              onChange={(e) => setQ({ ...q, part1: { ...q.part1, maxBt: Number(e.target.value) || 0 } })}
            />
          </label>
        </div>

        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <span className="text-[11px] uppercase tracking-[0.15em] text-gold">Part 2 — multiple choice</span>
            <button
              onClick={() =>
                setQ({ ...q, part2: [...q.part2, { prompt: '', options: ['', ''], correctAnswer: '', hfValue: 2, tier: 'medium' }] })
              }
              className="text-xs text-gold uppercase tracking-[0.15em]"
            >
              + Add question
            </button>
          </div>
          {q.part2.map((mc, i) => (
            <div key={i} className="rounded-md border border-blood/30 p-2 flex flex-col gap-2">
              <div className="flex items-center gap-2">
                <span className="text-gold text-xs w-4">{i + 1}</span>
                <input
                  type="number"
                  className="w-16 rounded-md bg-black/60 border border-blood/40 p-1.5 text-bone text-center text-sm"
                  value={mc.hfValue}
                  onChange={(e) => updMc(i, { hfValue: Number(e.target.value) || 0 })}
                />
                <span className="text-bone/40 text-xs">HF</span>
                <select
                  className="rounded-md bg-black/60 border border-blood/40 p-1.5 text-bone text-sm"
                  value={mc.tier}
                  onChange={(e) => updMc(i, { tier: e.target.value })}
                >
                  <option value="easy">easy</option>
                  <option value="medium">medium</option>
                  <option value="hard">hard</option>
                </select>
                <button
                  onClick={() => setQ({ ...q, part2: q.part2.filter((_, j) => j !== i) })}
                  className="ml-auto shrink-0 w-6 h-6 rounded-full border border-blood/50 text-blood-bright text-xs leading-none"
                  aria-label="Remove question"
                >
                  ×
                </button>
              </div>
              <textarea
                className={qInput}
                rows={2}
                placeholder="Question prompt"
                value={mc.prompt}
                onChange={(e) => updMc(i, { prompt: e.target.value })}
              />
              <div className="flex flex-col gap-1">
                {mc.options.map((opt, oi) => (
                  <div key={oi} className="flex items-center gap-2">
                    <input
                      type="radio"
                      name={`correct-${i}`}
                      checked={opt !== '' && mc.correctAnswer === opt}
                      onChange={() => updMc(i, { correctAnswer: opt })}
                      title="Mark correct answer"
                    />
                    <input
                      className={qInput}
                      placeholder={`Option ${oi + 1}`}
                      value={opt}
                      onChange={(e) => {
                        const options = mc.options.map((x, j) => (j === oi ? e.target.value : x));
                        const correctAnswer = mc.correctAnswer === opt ? e.target.value : mc.correctAnswer;
                        updMc(i, { options, correctAnswer });
                      }}
                    />
                    <button
                      onClick={() => updMc(i, { options: mc.options.filter((_, j) => j !== oi) })}
                      className="shrink-0 w-6 h-6 rounded-full border border-blood/50 text-blood-bright text-xs leading-none"
                      aria-label="Remove option"
                    >
                      ×
                    </button>
                  </div>
                ))}
                <button
                  onClick={() => updMc(i, { options: [...mc.options, ''] })}
                  className="text-xs text-gold uppercase tracking-[0.12em] self-start"
                >
                  + Add option
                </button>
              </div>
            </div>
          ))}
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={save}
            disabled={busy}
            className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            {busy ? 'Saving…' : 'Save questions'}
          </button>
          {note && <span className="text-bone/60 text-sm">{note}</span>}
        </div>
        <p className="text-[11px] text-bone/40">
          Saving replaces the question set for every Toast running this mystery and clears existing quiz answers. The
          numeric "Blood Tokens on hand" question is preserved automatically.
        </p>
      </div>
    </Card>
  );
};
