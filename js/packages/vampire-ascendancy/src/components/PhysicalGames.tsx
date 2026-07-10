import { useEffect, useState } from 'react';
import { getGames, getToken } from '../api';
import type { Game } from '../types';
import { accentFor, formatClock } from '../theme';
import { VampireMark } from './VampireMark';

const medal = ['🥇', '🥈', '🥉'];

// "Details of the physical games — order, what's been played, what's next, and
// winners." Reads the shared game list the GMs record results into.
export const PhysicalGames = () => {
  const [games, setGames] = useState<Game[] | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    let cancelled = false;
    const load = () => {
      getGames(token)
        .then((d) => !cancelled && setGames(d.games))
        .catch(() => {});
    };
    load();
    const id = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  // The next contest is the first one not yet played.
  const nextId = games?.find((g) => g.status !== 'played')?.id;

  return (
    <div className="pb-8">
      <header className="text-center mb-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold">Physical Games</p>
        <h1 className="mt-3 font-display text-3xl font-bold text-bone">The Night's Contests</h1>
      </header>

      {games === null ? (
        <p className="text-center text-bone/50">Gathering the contests…</p>
      ) : games.length === 0 ? (
        <div className="rounded-lg border border-blood/40 bg-black/40 p-8 text-center">
          <VampireMark className="w-12 h-12 mx-auto mb-3 opacity-80" />
          <p className="font-heading text-gold uppercase tracking-[0.3em] text-xs mb-2">Coming soon</p>
          <p className="text-bone/85">The evening's games will appear here as they're announced.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {games.map((g) => (
            <GameRow key={g.id} game={g} isNext={g.id === nextId} />
          ))}
        </div>
      )}
    </div>
  );
};

const GameRow = ({ game, isNext }: { game: Game; isNext: boolean }) => {
  const played = game.status === 'played';
  const places = [game.first, game.second, game.third];
  return (
    <div
      className={`rounded-lg border bg-black/40 p-4 ${
        isNext ? 'border-gold/60' : 'border-blood/30'
      }`}
    >
      <div className="flex items-center justify-between">
        <p className="text-bone font-semibold">
          {game.ordinal ? `${game.ordinal}. ` : ''}
          {game.name}
        </p>
        {played ? (
          <span className="text-xs uppercase tracking-[0.15em] text-green-400">Played</span>
        ) : isNext ? (
          <span className="text-xs uppercase tracking-[0.15em] text-gold">Up next</span>
        ) : (
          <span className="text-xs uppercase tracking-[0.15em] text-bone/40">Upcoming</span>
        )}
      </div>
      {game.startMinutes != null && game.endMinutes != null && (
        <p className="mt-1 text-xs text-bone/60">
          🕒 {formatClock(game.startMinutes)}–{formatClock(game.endMinutes)}
          {game.location && <span className="text-gold/80"> · 📍 {game.location}</span>}
        </p>
      )}
      {played && places.some((w) => w.length > 0) && (
        <div className="mt-2 flex flex-col gap-1">
          {places.map((winners, i) =>
            winners.length ? (
              <p key={i} className="text-sm text-bone/85">
                <span className="mr-2">{medal[i]}</span>
                {winners.map((w, j) => (
                  <span key={w.characterId}>
                    {w.characterName}
                    {w.house && (
                      <span className="ml-1" style={{ color: accentFor(w.house) }}>
                        · {w.house}
                      </span>
                    )}
                    {j < winners.length - 1 && <span className="text-bone/40">, </span>}
                  </span>
                ))}
              </p>
            ) : null
          )}
        </div>
      )}
    </div>
  );
};
