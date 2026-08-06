import { useEffect, useState } from 'react';
import { adminListHouses, adminUpdateHouse } from '../../superAdminApi';
import type { House } from '../../types';
import { Card } from '../gm/GameSection';

// House taglines are shared content (shown on every Toast's house pages).
export const SuperAdminHouses = () => {
  const [houses, setHouses] = useState<House[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminListHouses().then((d) => setHouses(d.houses)).finally(() => setLoading(false));
  }, []);

  if (loading) return <Card title="Houses">Loading…</Card>;

  return (
    <Card title="House taglines (shared — affects every Toast)">
      <div className="flex flex-col gap-2">
        {houses.map((h) => (
          <TaglineRow key={h.id} house={h} />
        ))}
      </div>
    </Card>
  );
};

const TaglineRow = ({ house }: { house: House }) => {
  const [tagline, setTagline] = useState(house.tagline ?? '');
  const [busy, setBusy] = useState(false);
  const dirty = tagline !== (house.tagline ?? '');

  const save = async () => {
    setBusy(true);
    try {
      await adminUpdateHouse(house.id, tagline);
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
