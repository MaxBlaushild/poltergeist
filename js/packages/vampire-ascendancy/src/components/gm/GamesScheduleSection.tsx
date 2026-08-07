import { useEffect, useState } from 'react';
import { gmListGames, gmCreateGame, gmUpdateGame, gmDeleteGame } from '../../gmApi';
import type { GMGame } from '../../gmApi';
import { Card } from './GameSection';
import { ScheduleCalendar } from './ScheduleCalendar';
import { ScheduleLine } from './GamesShared';

// Setup tab: stand up the night's contests and place them on the evening's
// timeline. Recording who won lives on the Scoring tab instead — this tab
// is done once the roster of games and their slots are set, before anyone's
// played anything.
export const GamesScheduleSection = () => {
  const [games, setGames] = useState<GMGame[]>([]);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [busy, setBusy] = useState(false);

  const load = () => {
    gmListGames()
      .then((g) => setGames(g.games))
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
        games.map((g) => <GameRow key={g.id} game={g} onChange={load} />)
      )}
    </div>
  );
};

const GameRow = ({ game, onChange }: { game: GMGame; onChange: () => void }) => {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState(game.name);
  const [busy, setBusy] = useState(false);
  const played = game.status === 'played';
  const title = `${game.ordinal ? game.ordinal + '. ' : ''}${game.name}`;

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

  return (
    <Card title={title}>
      <ScheduleLine game={game} />
      {played && (
        <p className="mb-2 text-xs text-green-400 uppercase tracking-[0.15em]">
          Recorded · scored on the Scoring tab
        </p>
      )}
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
          <button
            onClick={del}
            disabled={busy}
            className="py-2 rounded-md border border-blood/50 text-blood-bright uppercase tracking-[0.15em] text-xs disabled:opacity-40"
          >
            Delete game
          </button>
          {played && (
            <p className="text-[11px] text-bone/40">
              Rename is disabled while recorded — clear the result on the Scoring tab first.
            </p>
          )}
        </div>
      )}
    </Card>
  );
};
