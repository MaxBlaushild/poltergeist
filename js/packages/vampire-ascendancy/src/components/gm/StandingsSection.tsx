import { useEffect, useState } from 'react';
import { gmGetStandings, gmListHouses, gmUpdateHouse } from '../../gmApi';
import type { HouseStanding, House } from '../../types';
import { StandingsList } from '../Leaderboard';
import { Card } from './GameSection';

// The same house standings the players see, plus an editor for the house
// taglines. Polls the standings live.
export const StandingsSection = () => {
  const [standings, setStandings] = useState<HouseStanding[] | null>(null);
  const [houses, setHouses] = useState<House[]>([]);

  useEffect(() => {
    let cancelled = false;
    const load = () =>
      gmGetStandings()
        .then((d) => !cancelled && setStandings(d.standings))
        .catch(() => {});
    load();
    gmListHouses().then((d) => !cancelled && setHouses(d.houses)).catch(() => {});
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

      {houses.length > 0 && (
        <Card title="House taglines">
          <div className="flex flex-col gap-2">
            {houses.map((h) => (
              <TaglineRow key={h.id} house={h} />
            ))}
          </div>
        </Card>
      )}
    </div>
  );
};

const TaglineRow = ({ house }: { house: House }) => {
  const [tagline, setTagline] = useState(house.tagline ?? '');
  const [busy, setBusy] = useState(false);
  const dirty = tagline !== (house.tagline ?? '');

  const save = async () => {
    setBusy(true);
    try {
      await gmUpdateHouse(house.id, tagline);
      house.tagline = tagline; // reflect the saved value locally
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <span className="text-bone/70 text-sm w-40 shrink-0">House {house.name}</span>
      <input
        value={tagline}
        onChange={(e) => setTagline(e.target.value)}
        placeholder="Tagline"
        className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
      />
      {dirty && (
        <button
          onClick={save}
          disabled={busy}
          className="px-3 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-xs disabled:opacity-40"
        >
          Save
        </button>
      )}
    </div>
  );
};
