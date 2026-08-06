import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { bgiApi } from '../api/client';
import type { Game } from '../api/types';
import { getSessionId } from '../lib/session';

export default function Home() {
  const [games, setGames] = useState<Game[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    bgiApi
      .listGames()
      .then(setGames)
      .catch((e) => setError(String(e)));
  }, []);

  return (
    <div className="space-y-16">
      <section className="relative -mt-10 overflow-hidden bg-bgi-hero px-6 pb-6 pt-16 sm:px-12">
        <div className="relative max-w-xl">
          <span className="pill border-2 border-bgi-ink bg-bgi-glow text-bgi-ink">Printed to order · fit to your box</span>
          <h1 className="font-display mt-5 text-5xl font-bold leading-[1.05] text-bgi-ink sm:text-6xl">
            Tray sets that
            <br />
            <span className="relative inline-block">
              actually fit
              <svg aria-hidden="true" viewBox="0 0 300 20" className="absolute -bottom-2 left-0 h-4 w-full text-bgi-coral">
                <path d="M2 12 C 60 2, 120 18, 180 8 S 260 4, 298 10" fill="none" stroke="currentColor" strokeWidth="6" strokeLinecap="round" />
              </svg>
            </span>
          </h1>
          <p className="mb-8 mt-6 max-w-md text-lg text-bgi-ink/70">
            Sized to your sleeve class and your box — not a one-size insert. Pick your game below to start.
          </p>
        </div>
      </section>

      {error && <p className="text-red-600">{error}</p>}

      <section>
        <h2 className="font-display mb-5 text-2xl font-bold text-bgi-ink">Supported games</h2>
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3">
          {games.map((g) => (
            <GameCard key={g.slug} game={g} />
          ))}
        </div>
        <p className="mt-6 text-sm text-bgi-ink/60">
          Don't see your game? <WaitlistLink />
        </p>
      </section>
    </div>
  );
}

function GameCard({ game }: { game: Game }) {
  const href = game.productSlug ? `/configure/${game.slug}` : `/games/${game.slug}`;
  return (
    <Link
      to={href}
      onClick={() => bgiApi.recordEvent('game_selected', { sessionId: getSessionId(), gameSlug: game.slug })}
      className="card card-hover group block p-5"
    >
      <h3 className="font-display font-bold text-bgi-ink">{game.name}</h3>
      <p className="mt-1 text-sm text-bgi-ink/60">{game.publisher}{game.yearPublished ? ` · ${game.yearPublished}` : ''}</p>
      <p className="pill mt-3 border-2 border-bgi-ink/15 bg-white text-bgi-ink">built to order</p>
    </Link>
  );
}

function WaitlistLink() {
  const [email, setEmail] = useState('');
  const [requestedGame, setRequestedGame] = useState('');
  const [submitted, setSubmitted] = useState(false);

  if (submitted) return <span className="text-bgi-teal">Thanks — we'll let you know.</span>;

  return (
    <span className="inline-flex flex-wrap items-center gap-2 align-middle">
      <input
        type="text"
        placeholder="Which game?"
        value={requestedGame}
        onChange={(e) => setRequestedGame(e.target.value)}
        className="input-field inline w-40 py-1 text-sm"
      />
      <input
        type="email"
        placeholder="you@example.com"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        className="input-field inline w-48 py-1 text-sm"
      />
      <button
        className="btn-secondary py-1 text-sm"
        onClick={() => {
          if (!email || !requestedGame) return;
          bgiApi.submitWaitlist(email, requestedGame, getSessionId()).then(() => setSubmitted(true));
        }}
      >
        Request it
      </button>
    </span>
  );
}
