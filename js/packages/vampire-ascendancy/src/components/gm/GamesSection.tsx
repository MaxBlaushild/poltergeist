import { useEffect, useMemo, useState } from 'react';
import {
  gmListGames,
  gmCreateGame,
  gmRecordGameResult,
  gmListCharacters,
  gmUpdateGame,
  gmDeleteGame,
  gmClearGameResult,
} from '../../gmApi';
import type { GMGame, GMCharacter } from '../../gmApi';
import { Card } from './GameSection';
import { Combobox } from './Combobox';
import type { ComboOption } from './Combobox';
import { ScheduleCalendar } from './ScheduleCalendar';
import { formatClock } from '../../theme';

// GM Games tab: pre-seed / add the night's contests, then record the top three
// finishers as they emerge. The Blood Token / House Favor math is applied on the
// server (1st +5/+5, 2nd +3/+3, 3rd +1/+2, participation +1 BT).
export const GamesSection = () => {
  const [games, setGames] = useState<GMGame[]>([]);
  const [chars, setChars] = useState<GMCharacter[]>([]);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [busy, setBusy] = useState(false);

  const load = () => {
    Promise.all([gmListGames(), gmListCharacters()])
      .then(([g, c]) => {
        setGames(g.games);
        setChars(c.characters);
      })
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const addGame = async () => {
    const name = newName.trim();
    if (!name) return;
    setBusy(true);
    try {
      await gmCreateGame(name, games.length + 1);
      setNewName('');
      load();
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <p className="text-bone/50">Loading games…</p>;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-2">
        <input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && addGame()}
          placeholder="Add a game (e.g. Flip Cup)"
          className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
        />
        <button
          onClick={addGame}
          disabled={busy || !newName.trim()}
          className="px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          Add
        </button>
      </div>

      {games.length > 0 && <ScheduleCalendar games={games} onChange={load} />}

      {games.length === 0 ? (
        <p className="text-bone/50 text-sm">No games yet — add the night's contests above.</p>
      ) : (
        games.map((g) => <GameCard key={g.id} game={g} chars={chars} onChange={load} />)
      )}
    </div>
  );
};

const medal = ['🥇', '🥈', '🥉'];

const GameCard = ({
  game,
  chars,
  onChange,
}: {
  game: GMGame;
  chars: GMCharacter[];
  onChange: () => void;
}) => {
  const [first, setFirst] = useState<string[]>([]);
  const [second, setSecond] = useState<string[]>([]);
  const [third, setThird] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Sort options by house then name for easier scanning.
  const allOptions: ComboOption[] = useMemo(
    () =>
      [...chars]
        .sort(
          (a, b) => (a.house || '').localeCompare(b.house || '') || a.name.localeCompare(b.name)
        )
        .map((c) => ({ id: c.id, label: c.name, sub: c.house })),
    [chars]
  );

  const charHouse = useMemo(() => {
    const m: Record<string, string | undefined> = {};
    for (const c of chars) m[c.id] = c.house;
    return m;
  }, [chars]);
  const houseOf = (ids: string[]): string | undefined => {
    for (const id of ids) {
      const h = charHouse[id];
      if (h) return h;
    }
    return undefined;
  };

  // Options for a place, enforcing the house rules: everyone at a place must share
  // a house, and each place must be a different house from the others. So we hide
  // anyone already picked, anyone from a different house than this place's, and
  // anyone whose house is already claimed by another place.
  const optionsFor = (mine: string[], ...others: string[][]): ComboOption[] => {
    const taken = new Set([mine, ...others].flat());
    const myHouse = houseOf(mine);
    const otherHouses = new Set(others.map(houseOf).filter(Boolean) as string[]);
    return allOptions.filter((o) => {
      if (taken.has(o.id)) return false;
      const h = charHouse[o.id];
      if (myHouse && h !== myHouse) return false;
      if (h && otherHouses.has(h)) return false;
      return true;
    });
  };

  const anyPlaced = first.length + second.length + third.length > 0;

  const record = async () => {
    if (!anyPlaced) return;
    setBusy(true);
    setError(null);
    try {
      await gmRecordGameResult(game.id, { firstIds: first, secondIds: second, thirdIds: third });
      onChange();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not record the result.');
    } finally {
      setBusy(false);
    }
  };

  const title = `${game.ordinal ? game.ordinal + '. ' : ''}${game.name}`;

  if (game.status === 'played') {
    const places = [game.first, game.second, game.third];
    return (
      <Card title={title}>
        <ScheduleLine game={game} />
        <div className="flex flex-col gap-1.5">
          {places.map((winners, i) =>
            winners.length ? (
              <div key={i} className="flex items-start gap-2 text-bone">
                <span className="shrink-0">{medal[i]}</span>
                <span className="flex flex-wrap gap-x-2">
                  {winners.map((w, j) => (
                    <span key={w.characterId}>
                      {w.characterName}
                      {w.house && <span className="text-bone/50"> · {w.house}</span>}
                      {j < winners.length - 1 && <span className="text-bone/30">,</span>}
                    </span>
                  ))}
                </span>
              </div>
            ) : null
          )}
          <p className="mt-1 text-xs text-green-400 uppercase tracking-[0.15em]">Recorded · awards applied</p>
        </div>
        <ManageBar game={game} onChange={onChange} />
      </Card>
    );
  }

  return (
    <Card title={title}>
      <ScheduleLine game={game} />
      <div className="flex flex-col gap-3">
        <Field label="🥇 1st place">
          <Combobox
            options={optionsFor(first, second, third)}
            selected={first}
            onChange={setFirst}
            placeholder="Search a character…"
          />
        </Field>
        <Field label="🥈 2nd place">
          <Combobox
            options={optionsFor(second, first, third)}
            selected={second}
            onChange={setSecond}
            placeholder="Search a character…"
          />
        </Field>
        <Field label="🥉 3rd place">
          <Combobox
            options={optionsFor(third, first, second)}
            selected={third}
            onChange={setThird}
            placeholder="Search a character…"
          />
        </Field>

        <button
          onClick={record}
          disabled={busy || !anyPlaced}
          className="mt-1 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          Record result &amp; award House Favor
        </button>
        {error && <p className="text-sm text-blood-bright">{error}</p>}
        <p className="text-[11px] text-bone/40">
          House Favor is applied automatically — once per place, to that place's house: 1st +5, 2nd +3,
          3rd +2. Everyone sharing a place must be from the same house, and each place must be won by a
          different house. Blood Tokens are handed out in person: 1st +5, 2nd +3, 3rd +2, participants
          +1. Clear the result afterward to undo.
        </p>
      </div>
      <ManageBar game={game} onChange={onChange} />
    </Card>
  );
};

const Field = ({ label, children }: { label: string; children: React.ReactNode }) => (
  <div className="flex flex-col gap-1">
    <label className="text-xs text-bone/60">{label}</label>
    {children}
  </div>
);

const ScheduleLine = ({ game }: { game: GMGame }) =>
  game.startMinutes != null && game.endMinutes != null ? (
    <p className="text-xs text-bone/60 mb-2">
      🕒 {formatClock(game.startMinutes)}–{formatClock(game.endMinutes)}
      {game.location && <span className="text-gold/80"> · 📍 {game.location}</span>}
    </p>
  ) : null;

// Rename / delete / clear-result controls, collapsed by default. A recorded game
// can't be renamed (awards are matched by name) — clear it first.
const ManageBar = ({ game, onChange }: { game: GMGame; onChange: () => void }) => {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState(game.name);
  const [busy, setBusy] = useState(false);
  const played = game.status === 'played';

  const run = (fn: () => Promise<unknown>) => async () => {
    setBusy(true);
    try {
      await fn();
      onChange();
    } finally {
      setBusy(false);
    }
  };
  const rename = run(() => gmUpdateGame(game.id, name.trim(), game.ordinal));
  const del = () => {
    if (window.confirm(`Delete "${game.name}"?${played ? '\n\nThis also reverses the awards it applied.' : ''}`))
      run(() => gmDeleteGame(game.id))();
  };
  const clear = () => {
    if (window.confirm(`Clear the result for "${game.name}"?\n\nThis reverses the Blood Tokens and House Favor it awarded.`))
      run(() => gmClearGameResult(game.id))();
  };

  return (
    <div className="mt-3 pt-3 border-t border-blood/20">
      <button onClick={() => setOpen((o) => !o)} className="text-xs text-bone/50 uppercase tracking-[0.15em]">
        {open ? '▾ Manage' : '▸ Manage'}
      </button>
      {open && (
        <div className="mt-2 flex flex-col gap-2">
          {!played && (
            <div className="flex gap-2">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
              />
              <button
                onClick={rename}
                disabled={busy || !name.trim() || name.trim() === game.name}
                className="px-3 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-xs disabled:opacity-40"
              >
                Rename
              </button>
            </div>
          )}
          <div className="flex gap-2">
            {played && (
              <button
                onClick={clear}
                disabled={busy}
                className="flex-1 py-2 rounded-md border border-gold/50 text-gold uppercase tracking-[0.15em] text-xs disabled:opacity-40"
              >
                Clear result
              </button>
            )}
            <button
              onClick={del}
              disabled={busy}
              className="flex-1 py-2 rounded-md border border-blood/50 text-blood-bright uppercase tracking-[0.15em] text-xs disabled:opacity-40"
            >
              Delete game
            </button>
          </div>
          {played && (
            <p className="text-[11px] text-bone/40">Rename is disabled while recorded — clear the result first.</p>
          )}
        </div>
      )}
    </div>
  );
};

