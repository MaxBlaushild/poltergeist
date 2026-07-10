import { useEffect, useMemo, useState } from 'react';
import type { Game, HouseStanding } from '../types';
import { StandingsList } from './Leaderboard';
import { accentFor, formatClock } from '../theme';

const medal = ['🥇', '🥈', '🥉'];
const MAX_PER_PAGE = 7; // rows per carousel page — keeps everything on-screen (no scroll)
const PAGE_MS = 8000; // auto-advance interval

// The projector layout: house standings on the left, an auto-cycling board of the
// tournament games on the right so 13+ games fit a TV without scrolling. Purely
// presentational — the GM tab and the public /broadcast page both feed it data.
export const BroadcastView = ({
  standings,
  games,
  perPage,
}: {
  standings: HouseStanding[] | null;
  games: Game[] | null;
  perPage?: number;
}) => (
  <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
    <section className="lg:col-span-2">
      <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">House Standings</h2>
      {standings ? (
        <StandingsList standings={standings} linkHouses={false} />
      ) : (
        <p className="text-bone/50">Loading…</p>
      )}
    </section>
    <section className="lg:col-span-3">
      <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-3">Tournament Games</h2>
      {games ? <GamesCarousel games={games} perPage={perPage} /> : <p className="text-bone/50">Loading…</p>}
    </section>
  </div>
);

// Split into as few near-equal pages as possible (≤ perPage each).
const paginate = (games: Game[], perPageMax: number): Game[][] => {
  if (games.length === 0) return [];
  const numPages = Math.ceil(games.length / perPageMax);
  const per = Math.ceil(games.length / numPages);
  const pages: Game[][] = [];
  for (let i = 0; i < games.length; i += per) pages.push(games.slice(i, i + per));
  return pages;
};

const GamesCarousel = ({ games, perPage }: { games: Game[]; perPage?: number }) => {
  const pages = useMemo(() => paginate(games, perPage || MAX_PER_PAGE), [games, perPage]);
  const [page, setPage] = useState(0);
  const nextId = games.find((g) => g.status !== 'played')?.id;

  // Auto-advance through pages; reset when the page count changes.
  useEffect(() => {
    setPage((p) => (p >= pages.length ? 0 : p));
    if (pages.length <= 1) return;
    const id = setInterval(() => setPage((p) => (p + 1) % pages.length), PAGE_MS);
    return () => clearInterval(id);
  }, [pages.length]);

  if (games.length === 0) return <p className="text-bone/50">No games yet.</p>;
  const current = pages[Math.min(page, pages.length - 1)] ?? [];

  return (
    <div>
      <div key={page} className="flex flex-col gap-2 animate-[fade_0.5s_ease]">
        {current.map((g) => (
          <GameRow key={g.id} game={g} isNext={g.id === nextId} />
        ))}
      </div>
      {pages.length > 1 && (
        <div className="flex items-center justify-center gap-2 mt-4">
          {pages.map((_, i) => (
            <span
              key={i}
              className={`h-1.5 rounded-full transition-all ${
                i === page ? 'w-6 bg-gold' : 'w-1.5 bg-bone/25'
              }`}
            />
          ))}
        </div>
      )}
    </div>
  );
};

const GameRow = ({ game: g, isNext }: { game: Game; isNext: boolean }) => {
  const places = [g.first, g.second, g.third];
  return (
    <div
      className={`rounded-lg border bg-black/40 px-3 py-2 ${
        isNext ? 'border-gold/60' : 'border-blood/30'
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-bone font-semibold truncate">
          {g.ordinal ? `${g.ordinal}. ` : ''}
          {g.name}
        </p>
        <span className="shrink-0 flex items-center gap-2 text-xs">
          {g.startMinutes != null && g.endMinutes != null && (
            <span className="text-bone/60 whitespace-nowrap">
              🕒 {formatClock(g.startMinutes)}
              {g.location && <span className="text-gold/80"> · 📍 {g.location}</span>}
            </span>
          )}
          {g.status === 'played' ? (
            <span className="uppercase tracking-[0.15em] text-green-400">Played</span>
          ) : isNext ? (
            <span className="uppercase tracking-[0.15em] text-gold">Up next</span>
          ) : (
            <span className="uppercase tracking-[0.15em] text-bone/40">Upcoming</span>
          )}
        </span>
      </div>
      {g.status === 'played' && places.some((w) => w.length > 0) && (
        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-0.5">
          {places.map((winners, i) =>
            winners.length ? (
              <p key={i} className="text-sm text-bone/85">
                <span className="mr-1">{medal[i]}</span>
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
