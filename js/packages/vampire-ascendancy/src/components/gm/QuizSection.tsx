import { useEffect, useState } from 'react';
import {
  gmSetPart1Open,
  gmSetPart2Open,
  gmGradePart1,
  gmOverridePart1BT,
  gmRescorePart2,
  gmListQuizSubmissions,
  gmGetStandings,
  gmGetQuizQuestions,
  gmUpdateQuizQuestions,
} from '../../gmApi';
import type { GMQuizSubmission, GMQuizQuestions } from '../../gmApi';
import type { GameState, HouseStanding } from '../../types';
import { accentFor, formatHF, houseLabel } from '../../theme';
import { Card } from './GameSection';

export const QuizSection = ({
  state,
  onChange,
}: {
  state: GameState | null;
  onChange: () => void;
}) => {
  const [subs, setSubs] = useState<GMQuizSubmission[]>([]);
  const [standings, setStandings] = useState<HouseStanding[]>([]);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  const loadSubs = () => {
    gmListQuizSubmissions().then((d) => setSubs(d.submissions || [])).catch(() => {});
    gmGetStandings().then((d) => setStandings(d.standings || [])).catch(() => {});
  };
  useEffect(() => {
    loadSubs();
    const id = setInterval(loadSubs, 6000);
    return () => clearInterval(id);
  }, []);

  const part1Open = !!state?.quizPart1Open;
  const part2Open = !!state?.quizPart2Open;

  const wrap = async (fn: () => Promise<unknown>, msg?: string) => {
    setBusy(true);
    setNote(null);
    try {
      await fn();
      if (msg) setNote(msg);
      onChange();
      loadSubs();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Failed.');
    } finally {
      setBusy(false);
    }
  };

  const part1Subs = subs.filter((s) => s.part === 1);
  const part2Subs = subs.filter((s) => s.part === 2);

  return (
    <div className="flex flex-col gap-6">
      <QuizEditor />

      {/* ---- Part 1 ---- */}
      <Card title="Part 1 — Open-end → Blood Tokens">
        <div className="flex items-center justify-between gap-3">
          <p className="text-bone text-sm">
            <span className={part1Open ? 'text-green-400' : 'text-blood-bright'}>
              {part1Open ? 'OPEN' : 'CLOSED'}
            </span>{' '}
            · timed open-end, AI-graded.
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => wrap(() => gmSetPart1Open(!part1Open))}
              disabled={busy}
              className={`px-4 py-2 rounded-md text-sm uppercase tracking-[0.12em] disabled:opacity-40 ${
                part1Open ? 'border border-blood/50 text-bone/70' : 'bg-blood text-bone'
              }`}
            >
              {part1Open ? 'Close' : 'Open'}
            </button>
            <button
              onClick={() => wrap(() => gmGradePart1(), 'Grading started — scores will appear shortly.')}
              disabled={busy}
              className="px-4 py-2 rounded-md text-sm uppercase tracking-[0.12em] border border-gold/50 text-gold disabled:opacity-40"
            >
              Grade
            </button>
          </div>
        </div>
        {note && <p className="text-bone/60 text-sm mt-2">{note}</p>}
      </Card>

      {part1Subs.map((s) => (
        <Part1Row key={s.id} sub={s} onSaved={loadSubs} />
      ))}

      {/* ---- Part 2 ---- */}
      <Card title="Part 2 — Multiple choice → House Favor">
        <div className="flex items-center justify-between gap-3">
          <p className="text-bone text-sm">
            <span className={part2Open ? 'text-green-400' : 'text-blood-bright'}>
              {part2Open ? 'OPEN' : 'CLOSED'}
            </span>{' '}
            · closing it scores per house.
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => wrap(() => gmSetPart2Open(!part2Open))}
              disabled={busy}
              className={`px-4 py-2 rounded-md text-sm uppercase tracking-[0.12em] disabled:opacity-40 ${
                part2Open ? 'border border-blood/50 text-bone/70' : 'bg-blood text-bone'
              }`}
            >
              {part2Open ? 'Close & score' : 'Open'}
            </button>
            <button
              onClick={() => wrap(() => gmRescorePart2(), 'Re-scored.')}
              disabled={busy}
              className="px-4 py-2 rounded-md text-sm uppercase tracking-[0.12em] border border-gold/50 text-gold disabled:opacity-40"
            >
              Rescore
            </button>
          </div>
        </div>
        {part1Open && (
          <p className="text-bone/50 text-xs mt-2">
            Part 1 is still open — you can open Part 2 now and keep reviewing Part 1 scores.
          </p>
        )}
      </Card>

      <QuizResults subs={subs} standings={standings} />
      <Part2Summary subs={part2Subs} />
    </div>
  );
};

// Final results: the winning house is decided by cumulative House Favor (favor
// before the quiz + quiz favor = the current standings). The winning player is
// the highest Blood Token holder within that house — quiz BT (Part 1) plus the
// physical count each player self-reported in the numeric Part 2 question.
type PlayerRow = {
  character: string;
  house: string;
  correct: number;
  mcTotal: number;
  quizBt: number;
  physicalBt: number;
  total: number;
};

const buildPlayerRows = (subs: GMQuizSubmission[]): PlayerRow[] => {
  const byChar = new Map<string, PlayerRow>();
  const row = (name: string, house: string) => {
    const key = name || '—';
    if (!byChar.has(key))
      byChar.set(key, {
        character: key,
        house,
        correct: 0,
        mcTotal: 0,
        quizBt: 0,
        physicalBt: 0,
        total: 0,
      });
    const r = byChar.get(key)!;
    if (house) r.house = house;
    return r;
  };
  for (const s of subs) {
    const r = row(s.characterName, s.houseName);
    if (s.part === 1) {
      r.quizBt = s.awardedBt || 0;
    } else if (s.part === 2) {
      if (s.questionType === 'number') {
        r.physicalBt = parseInt((s.answer || '').replace(/[^0-9]/g, ''), 10) || 0;
      } else {
        r.mcTotal += 1;
        if (s.isCorrect) r.correct += 1;
      }
    }
  }
  const rows = [...byChar.values()];
  rows.forEach((r) => (r.total = r.quizBt + r.physicalBt));
  return rows;
};

const QuizResults = ({
  subs,
  standings,
}: {
  subs: GMQuizSubmission[];
  standings: HouseStanding[];
}) => {
  if (subs.length === 0) return null;
  const houses = [...standings].sort((a, b) => b.favor - a.favor);
  const winningHouse = houses[0]?.name;
  const players = buildPlayerRows(subs).sort(
    (a, b) => a.house.localeCompare(b.house) || b.total - a.total
  );
  const winner = players
    .filter((p) => p.house === winningHouse)
    .sort((a, b) => b.total - a.total || b.quizBt - a.quizBt || b.correct - a.correct)[0];

  return (
    <div className="flex flex-col gap-4">
      {winningHouse && (
        <Card title="The Throne">
          <p className="text-bone">
            Winning house:{' '}
            <span className="font-semibold" style={{ color: accentFor(winningHouse) }}>
              {houseLabel(winningHouse)}
            </span>{' '}
            <span className="text-bone/50">({formatHF(houses[0].favor)} favor)</span>
          </p>
          {winner && (
            <p className="text-bone mt-1">
              Throne:{' '}
              <span className="text-gold font-semibold">{winner.character}</span>{' '}
              <span className="text-bone/50">
                — {winner.total} BT ({winner.quizBt} quiz + {winner.physicalBt} on hand)
              </span>
            </p>
          )}
        </Card>
      )}

      <Card title="House results (cumulative favor)">
        <Table
          head={['#', 'House', 'Favor']}
          rows={houses.map((h, i) => [
            String(i + 1),
            houseLabel(h.name),
            formatHF(h.favor),
            h.name === winningHouse ? 'win' : '',
          ])}
        />
      </Card>

      <Card title="Player results">
        <Table
          head={['Character', 'House', 'MC', 'Quiz BT', 'On hand', 'Total BT']}
          rows={players.map((p) => [
            p.character,
            p.house,
            `${p.correct}/${p.mcTotal}`,
            String(p.quizBt),
            String(p.physicalBt),
            String(p.total),
            p.house === winningHouse && winner && p.character === winner.character ? 'win' : '',
          ])}
        />
      </Card>
    </div>
  );
};

// Simple table; a trailing 'win' marker cell highlights the winning row.
const Table = ({ head, rows }: { head: string[]; rows: string[][] }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-sm border-collapse">
      <thead>
        <tr>
          {head.map((h, i) => (
            <th
              key={i}
              className={`border-b border-blood/40 pb-2 font-heading uppercase tracking-[0.12em] text-[11px] text-gold ${
                i === 0 ? 'text-left' : 'text-right'
              }`}
            >
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((r, ri) => {
          const win = r[r.length - 1] === 'win';
          const cells = r.slice(0, head.length);
          return (
            <tr key={ri} className={`border-b border-blood/15 last:border-0 ${win ? 'bg-gold/10' : ''}`}>
              {cells.map((c, ci) => (
                <td
                  key={ci}
                  className={`py-2 ${ci === 0 ? 'text-left text-bone' : 'text-right text-bone/80'}`}
                >
                  {c}
                  {win && ci === 0 && <span className="ml-2 text-gold">👑</span>}
                </td>
              ))}
            </tr>
          );
        })}
      </tbody>
    </table>
  </div>
);

const Part1Row = ({ sub, onSaved }: { sub: GMQuizSubmission; onSaved: () => void }) => {
  const [bt, setBt] = useState(String(sub.awardedBt));
  const [busy, setBusy] = useState(false);

  // Pre-fill the BT box with the AI's recommendation when grading lands, so the
  // GM just confirms. Keyed on aiScore so a later poll doesn't clobber GM edits.
  useEffect(() => {
    setBt(String(sub.awardedBt));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sub.aiScore]);

  const save = async () => {
    setBusy(true);
    try {
      await gmOverridePart1BT(sub.id, Number(bt) || 0);
      onSaved();
    } finally {
      setBusy(false);
    }
  };
  return (
    <Card title={`${sub.characterName || '—'} · ${sub.houseName}`}>
      <p className="text-bone bg-black/50 rounded-md p-3 mb-2 whitespace-pre-wrap text-sm">
        {sub.answer || <span className="text-bone/40">— no answer —</span>}
      </p>
      {sub.aiRationale && (
        <p className="text-xs text-bone/60 italic mb-3">AI: “{sub.aiRationale}”</p>
      )}
      <div className="flex items-center gap-2">
        <span className="text-xs text-bone/50">
          AI score: {sub.aiScore == null ? '—' : sub.aiScore}
        </span>
        <label className="text-xs text-bone/50 ml-auto">BT</label>
        <input
          value={bt}
          onChange={(e) => setBt(e.target.value.replace(/[^0-9]/g, ''))}
          className="w-16 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-center"
        />
        <button
          onClick={save}
          disabled={busy}
          className="px-4 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.12em] text-sm disabled:opacity-40"
        >
          Confirm
        </button>
      </div>
    </Card>
  );
};

const Part2Summary = ({ subs }: { subs: GMQuizSubmission[] }) => {
  if (subs.length === 0) return <p className="text-bone/50">No Part 2 answers yet.</p>;
  // Group by question: correct / answered.
  const byQ = new Map<number, { prompt: string; correct: number; total: number }>();
  for (const s of subs) {
    if (!byQ.has(s.ordinal)) byQ.set(s.ordinal, { prompt: s.prompt, correct: 0, total: 0 });
    const g = byQ.get(s.ordinal)!;
    g.total += 1;
    if (s.isCorrect) g.correct += 1;
  }
  const groups = [...byQ.entries()].sort((a, b) => a[0] - b[0]);
  return (
    <Card title="Part 2 results">
      <div className="flex flex-col gap-2 text-sm">
        {groups.map(([ord, g]) => (
          <div key={ord} className="flex items-center gap-3">
            <span className="text-bone/50 w-6">{ord}.</span>
            <span className="text-bone flex-1">{g.prompt}</span>
            <span className="text-bone/70">
              {g.correct}/{g.total} correct
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
};

// Collapsible editor for the quiz content: the Part 1 open-end prompt + rubric,
// and the Part 2 multiple-choice questions. Saving replaces the question set (and
// clears any existing answers), so it's a pre-quiz operation.
const qInput = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

const QuizEditor = () => {
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState<GMQuizQuestions | null>(null);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    if (open && !q) gmGetQuizQuestions().then(setQ).catch(() => setNote('Could not load questions.'));
  }, [open, q]);

  const updMc = (i: number, patch: Partial<GMQuizQuestions['part2'][number]>) =>
    setQ((prev) => (prev ? { ...prev, part2: prev.part2.map((x, j) => (j === i ? { ...x, ...patch } : x)) } : prev));

  const save = async () => {
    if (!q) return;
    if (!window.confirm('Save quiz questions?\n\nThis replaces the question set and clears any existing quiz answers. Do this before the quiz runs.'))
      return;
    setBusy(true);
    setNote(null);
    try {
      await gmUpdateQuizQuestions(q);
      setNote('Saved.');
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title="Edit questions">
      <button onClick={() => setOpen((o) => !o)} className="text-xs text-gold uppercase tracking-[0.15em]">
        {open ? '▾ Hide editor' : '▸ Edit quiz questions'}
      </button>
      {open && !q && <p className="text-bone/50 text-sm mt-2">{note || 'Loading…'}</p>}
      {open && q && (
        <div className="mt-3 flex flex-col gap-5">
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
                  setQ({ ...q, part2: [...q.part2, { prompt: '', options: ['', ''], correctAnswer: '', hfValue: 3, tier: 'medium' }] })
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
            Saving replaces the question set and clears existing quiz answers. The numeric "Blood
            Tokens on hand" question is preserved automatically.
          </p>
        </div>
      )}
    </Card>
  );
};
