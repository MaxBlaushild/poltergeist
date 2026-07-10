import { useEffect, useState } from 'react';
import { gmGetStandings, gmListGames } from '../../gmApi';
import type { GMGame } from '../../gmApi';
import type { HouseStanding } from '../../types';
import { StandingsList } from '../Leaderboard';
import { accentFor, formatClock } from '../../theme';

const medal = ['🥇', '🥈', '🥉'];

// The projector view: house standings and the games board side by side, both
// live-polling, for casting to a TV during play.
export const BroadcastScreen = () => {
  const [standings, setStandings] = useState<HouseStanding[] | null>(null);
  const [games, setGames] = useState<GMGame[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = () => {
      gmGetStandings().then((d) => !cancelled && setStandings(d.standings)).catch(() => {});
      gmListGames().then((d) => !cancelled && setGames(d.games)).catch(() => {});
    };
    load();
    const id = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  return (
    <div>
      <p className="text-bone/50 text-sm mb-4">
        Cast this to the TV — House Favor and game results update live.
      </p>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <section>
          <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">
            House Standings
          </h2>
          {standings ? (
            <StandingsList standings={standings} linkHouses={false} />
          ) : (
            <p className="text-bone/50">Loading…</p>
          )}
        </section>
        <section>
          <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">
            Physical Games
          </h2>
          {games ? <GamesBoard games={games} /> : <p className="text-bone/50">Loading…</p>}
        </section>
      </div>
    </div>
  );
};

const GamesBoard = ({ games }: { games: GMGame[] }) => {
  if (games.length === 0) return <p className="text-bone/50">No games yet.</p>;
  const nextId = games.find((g) => g.status !== 'played')?.id;
  return (
    <div className="flex flex-col gap-2">
      {games.map((g) => {
        const places = [g.first, g.second, g.third];
        const isNext = g.id === nextId;
        return (
          <div
            key={g.id}
            className={`rounded-lg border bg-black/40 p-3 ${isNext ? 'border-gold/60' : 'border-blood/30'}`}
          >
            <div className="flex items-center justify-between">
              <p className="text-bone font-semibold">
                {g.ordinal ? `${g.ordinal}. ` : ''}
                {g.name}
              </p>
              {g.status === 'played' ? (
                <span className="text-xs uppercase tracking-[0.15em] text-green-400">Played</span>
              ) : isNext ? (
                <span className="text-xs uppercase tracking-[0.15em] text-gold">Up next</span>
              ) : (
                <span className="text-xs uppercase tracking-[0.15em] text-bone/40">Upcoming</span>
              )}
            </div>
            {g.startMinutes != null && g.endMinutes != null && (
              <p className="mt-0.5 text-xs text-bone/60">
                🕒 {formatClock(g.startMinutes)}–{formatClock(g.endMinutes)}
                {g.location && <span className="text-gold/80"> · 📍 {g.location}</span>}
              </p>
            )}
            {g.status === 'played' && places.some((w) => w.length > 0) && (
              <div className="mt-1 flex flex-col gap-0.5">
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
      })}
    </div>
  );
};
