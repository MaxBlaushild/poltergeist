import { useEffect, useState } from 'react';
import { gmGetStandings } from '../../gmApi';
import type { HouseStanding } from '../../types';
import { StandingsList } from '../Leaderboard';

// The same house standings the players see, for the GM console. Polls live.
export const StandingsSection = () => {
  const [standings, setStandings] = useState<HouseStanding[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = () =>
      gmGetStandings()
        .then((d) => !cancelled && setStandings(d.standings))
        .catch(() => {});
    load();
    const id = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  if (!standings) return <p className="text-bone/50">Tallying the favor…</p>;
  return <StandingsList standings={standings} linkHouses={false} />;
};
