import { useEffect, useState } from 'react';
import { gmGetStandings, gmListGames } from '../../gmApi';
import type { GMGame } from '../../gmApi';
import type { HouseStanding } from '../../types';
import { BroadcastView } from '../BroadcastView';

// GM console tab: an in-console preview of the projector view, plus a link to the
// standalone, no-auth /broadcast page for casting to a TV.
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
      <div className="flex items-center justify-between gap-3 mb-4">
        <p className="text-bone/50 text-sm">Live House Favor + game results.</p>
        <a
          href="/broadcast"
          target="_blank"
          rel="noreferrer"
          className="shrink-0 px-4 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-xs hover:bg-blood-bright"
        >
          Open projector screen ↗
        </a>
      </div>
      <BroadcastView standings={standings} games={games} />
    </div>
  );
};
