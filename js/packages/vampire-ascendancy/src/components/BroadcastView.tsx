import { useEffect, useMemo, useState } from 'react';
import type { Game, HouseStanding } from '../types';
import { StandingsList } from './Leaderboard';
import { accentFor, formatClock } from '../theme';

const medal = ['🥇', '🥈', '🥉'];
const MAX_PER_PAGE = 7; // rows per carousel page — keeps everything on-screen (no scroll)
const PAGE_MS = 8000; // auto-advance interval

const scheduled = (g: Game) => g.startMinutes != null && g.endMinutes != null;
const byTime = (a: Game, b: Game) =>
  (a.startMinutes ?? 100000) - (b.startMinutes ?? 100000) || a.ordinal - b.ordinal;
const nowMinutes = () => {
  const d = new Date();
  return d.getHours() * 60 + d.getMinutes();
};

// The projector layout: house standings on the left, the tournament games on the
// right — a "Happening Now / Up Next" spotlight over the full time-ordered board.
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
      {games ? <TournamentGames games={games} perPage={perPage} /> : <p className="text-bone/50">Loading…</p>}
    </section>
  </div>
);

const TournamentGames = ({ games, perPage }: { games: Game[]; perPage?: number }) => {
  const ordered = useMemo(() => [...games].sort(byTime), [games]);
  const now = nowMinutes();

  // "Now" = the next unplayed game whose scheduled window contains the current
  // time. "Next" = the following unplayed game (or the first, if none is current).
  const pending = ordered.filter((g) => g.status !== 'played');
  const current = pending.find((g) => scheduled(g) && g.startMinutes! <= now && now < g.endMinutes!);
  const next = pending.find((g) => g.id !== current?.id);

  const spot: { game: Game; label: string; tone: 'now' | 'next' }[] = [];
  if (current) spot.push({ game: current, label: 'Happening Now', tone: 'now' });
  if (next) spot.push({ game: next, label: 'Up Next', tone: 'next' });
  const spotIds = new Set(spot.map((s) => s.game.id));
  const rest = ordered.filter((g) => !spotIds.has(g.id));

  if (games.length === 0) return <p className="text-bone/50">No games yet.</p>;

  return (
    <div>
      {spot.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-5">
          {spot.map(({ game, label, tone }) => (
            <SpotlightCard key={game.id} game={game} label={label} tone={tone} />
          ))}
        </div>
      )}
      {rest.length > 0 && (
        <>
          {spot.length > 0 && (
            <h3 className="text-[11px] uppercase tracking-[0.25em] text-bone/40 mb-2">Full schedule</h3>
          )}
          <GamesCarousel games={rest} perPage={perPage} />
        </>
      )}
    </div>
  );
};

const SpotlightCard = ({
  game: g,
  label,
  tone,
}: {
  game: Game;
  label: string;
  tone: 'now' | 'next';
}) => (
  <div
    className={`rounded-lg border-2 p-3 ${
      tone === 'now' ? 'border-green-500/70 bg-green-900/20' : 'border-gold/70 bg-gold/5'
    }`}
  >
    <p
      className={`font-heading text-xs uppercase tracking-[0.3em] ${
        tone === 'now' ? 'text-green-400' : 'text-gold'
      }`}
    >
      {label}
    </p>
    <p className="text-bone font-bold text-lg leading-tight mt-1">
      {g.ordinal ? `${g.ordinal}. ` : ''}
      {g.name}
    </p>
    {scheduled(g) && (
      <p className="text-sm text-bone/70 mt-1">
        🕒 {formatClock(g.startMinutes!)}–{formatClock(g.endMinutes!)}
        {g.location && <span className="text-gold/80"> · 📍 {g.location}</span>}
      </p>
    )}
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

  // Auto-advance through pages; reset when the page count changes.
  useEffect(() => {
    setPage((p) => (p >= pages.length ? 0 : p));
    if (pages.length <= 1) return;
    const id = setInterval(() => setPage((p) => (p + 1) % pages.length), PAGE_MS);
    return () => clearInterval(id);
  }, [pages.length]);

  if (games.length === 0) return null;
  const current = pages[Math.min(page, pages.length - 1)] ?? [];

  return (
    <div>
      <div key={page} className="flex flex-col gap-2 animate-[fade_0.5s_ease]">
        {current.map((g) => (
          <GameRow key={g.id} game={g} />
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

const GameRow = ({ game: g }: { game: Game }) => {
  const places = [g.first, g.second, g.third];
  return (
    <div className="rounded-lg border border-blood/30 bg-black/40 px-3 py-2">
      <div className="flex items-center justify-between gap-2">
        <p className="text-bone font-semibold truncate">
          {g.ordinal ? `${g.ordinal}. ` : ''}
          {g.name}
        </p>
        <span className="shrink-0 flex items-center gap-2 text-xs">
          {scheduled(g) && (
            <span className="text-bone/60 whitespace-nowrap">
              🕒 {formatClock(g.startMinutes!)}
              {g.location && <span className="text-gold/80"> · 📍 {g.location}</span>}
            </span>
          )}
          {g.status === 'played' ? (
            <span className="uppercase tracking-[0.15em] text-green-400">Played</span>
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
