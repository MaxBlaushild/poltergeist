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
  const [first, setFirst] = useState('');
  const [second, setSecond] = useState('');
  const [third, setThird] = useState('');
  const [participants, setParticipants] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);

  // Sort options by house then name for easier scanning.
  const options = useMemo(
    () =>
      [...chars].sort((a, b) =>
        (a.house || '').localeCompare(b.house || '') || a.name.localeCompare(b.name)
      ),
    [chars]
  );

  const record = async () => {
    if (!first && !second && !third) return;
    setBusy(true);
    try {
      await gmRecordGameResult(game.id, {
        firstId: first || undefined,
        secondId: second || undefined,
        thirdId: third || undefined,
        participantIds: participants,
      });
      onChange();
    } finally {
      setBusy(false);
    }
  };

  const title = `${game.ordinal ? game.ordinal + '. ' : ''}${game.name}`;

  if (game.status === 'played') {
    const places = [game.first, game.second, game.third];
    return (
      <Card title={title}>
        <div className="flex flex-col gap-1">
          {places.map((p, i) =>
            p ? (
              <p key={i} className="text-bone">
                <span className="mr-2">{medal[i]}</span>
                {p.characterName}
                {p.house && <span className="text-bone/50"> · {p.house}</span>}
              </p>
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
      <div className="flex flex-col gap-2">
        <WinnerSelect label="🥇 1st" value={first} onChange={setFirst} options={options} />
        <WinnerSelect label="🥈 2nd" value={second} onChange={setSecond} options={options} />
        <WinnerSelect label="🥉 3rd" value={third} onChange={setThird} options={options} />

        <label className="text-xs text-bone/50 mt-1">Other participants (+1 BT each)</label>
        <select
          multiple
          size={5}
          value={participants}
          onChange={(e) =>
            setParticipants(Array.from(e.target.selectedOptions, (o) => o.value))
          }
          className="rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        >
          {options.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
              {c.house ? ` (${c.house})` : ''}
            </option>
          ))}
        </select>

        <button
          onClick={record}
          disabled={busy || (!first && !second && !third)}
          className="mt-2 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          Record result &amp; award
        </button>
        <p className="text-[11px] text-bone/40">
          BT/HF are applied automatically: 1st +5/+5, 2nd +3/+3, 3rd +1/+2. You can clear the result
          afterward to undo the awards.
        </p>
      </div>
      <ManageBar game={game} onChange={onChange} />
    </Card>
  );
};

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

const WinnerSelect = ({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: GMCharacter[];
}) => (
  <div className="flex items-center gap-2">
    <span className="w-14 text-sm text-bone/70 shrink-0">{label}</span>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2 text-bone"
    >
      <option value="">— none —</option>
      {options.map((c) => (
        <option key={c.id} value={c.id}>
          {c.name}
          {c.house ? ` (${c.house})` : ''}
        </option>
      ))}
    </select>
  </div>
);
