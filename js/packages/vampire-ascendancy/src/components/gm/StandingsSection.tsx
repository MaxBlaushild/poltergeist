import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { gmGetStandings } from '../../gmApi';
import type { HouseStanding } from '../../types';
import { StandingsList } from '../Leaderboard';

// The same house standings the players see. House taglines are shared
// content, edited in the Super Admin dashboard, not here. Polls the
// standings live.
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

  return (
    <div className="flex flex-col gap-4">
      {!standings ? (
        <p className="text-bone/50">Tallying the favor…</p>
      ) : (
        <StandingsList standings={standings} linkHouses={false} />
      )}
      <p className="text-bone/40 text-xs">
        Edit house taglines in the{' '}
        <Link to="/admin" className="text-gold underline underline-offset-2">
          Super Admin dashboard
        </Link>
        .
      </p>
    </div>
  );
};
