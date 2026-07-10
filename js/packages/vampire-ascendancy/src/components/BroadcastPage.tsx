import { useEffect, useState } from 'react';
import { getBroadcastStandings, getBroadcastGames } from '../api';
import type { Game, HouseStanding } from '../types';
import { BroadcastView } from './BroadcastView';
import { VampireMark } from './VampireMark';

// Standalone, no-auth projector page (route: /broadcast). Cast this to a TV — it
// polls the public standings + games feed and never needs a login.
export const BroadcastPage = () => {
  const [standings, setStandings] = useState<HouseStanding[] | null>(null);
  const [games, setGames] = useState<Game[] | null>(null);

  // Sized for a big TV. Tune live on the screen via URL params, e.g.
  // /broadcast?scale=1.7&perpage=4 — scale zooms everything, perpage sets how many
  // games show per carousel page.
  const params = new URLSearchParams(window.location.search);
  const scale = Math.min(3, Math.max(1, Number(params.get('scale')) || 1.5));
  const perPage = Math.max(1, Number(params.get('perpage')) || 5);

  // Zoom the whole page by bumping the root font size (Tailwind sizes are rem-based,
  // so text and spacing scale together). Reset when leaving the projector page.
  useEffect(() => {
    const prev = document.documentElement.style.fontSize;
    document.documentElement.style.fontSize = `${Math.round(scale * 100)}%`;
    return () => {
      document.documentElement.style.fontSize = prev;
    };
  }, [scale]);

  useEffect(() => {
    let cancelled = false;
    const load = () => {
      getBroadcastStandings().then((d) => !cancelled && setStandings(d.standings)).catch(() => {});
      getBroadcastGames().then((d) => !cancelled && setGames(d.games)).catch(() => {});
    };
    load();
    const id = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  return (
    <div className="min-h-screen bg-blood-ink px-6 py-8">
      <header className="text-center mb-8">
        <VampireMark className="w-10 h-10 mx-auto mb-2" />
        <p className="text-xs uppercase tracking-[0.4em] text-gold">The Crimson Toast</p>
        <h1 className="mt-1 font-display text-4xl md:text-5xl font-bold text-bone">Favor of the Court</h1>
      </header>
      <div className="max-w-6xl mx-auto">
        <BroadcastView standings={standings} games={games} perPage={perPage} />
      </div>
    </div>
  );
};
