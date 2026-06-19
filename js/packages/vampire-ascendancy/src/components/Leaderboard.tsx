import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getLeaderboard, getToken } from '../api';
import type { HouseStanding } from '../types';
import { accentFor, houseInfoFor, formatHF } from '../theme';

// Small house emblem; source art is black-on-white so we invert it. Hides if
// the image hasn't been added yet.
const Sigil = ({ src }: { src: string }) => {
  const [failed, setFailed] = useState(false);
  if (failed) return null;
  return (
    <img
      src={src}
      alt=""
      onError={() => setFailed(true)}
      className="w-8 h-8 object-contain"
      style={{ filter: 'invert(1)', mixBlendMode: 'screen' }}
    />
  );
};

export const Leaderboard = ({ myHouse }: { myHouse?: string }) => {
  const [standings, setStandings] = useState<HouseStanding[] | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    let cancelled = false;
    const load = () => {
      getLeaderboard(token)
        .then((d) => !cancelled && setStandings(d.standings))
        .catch(() => {
          /* keep last good standings on a flaky poll */
        });
    };
    load();
    const id = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  return (
    <div className="pt-2 pb-8">
      <header className="text-center mb-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold">House Standings</p>
        <h1 className="mt-3 font-display text-3xl font-bold text-bone">The Crimson Toast</h1>
        <p className="mt-1 text-bone/70 text-sm">Favor of the Court</p>
      </header>

      {!standings ? (
        <p className="text-center text-bone/50">Tallying the favor…</p>
      ) : (
        <div className="flex flex-col gap-3">
          {standings.map((h, i) => {
            const accent = accentFor(h.name);
            const mine = myHouse === h.name;
            const sigil = houseInfoFor(h.name).sigil;
            return (
              <Link
                key={h.houseId}
                to={`/house/${h.houseId}`}
                className={`flex items-center gap-3 rounded-lg border bg-black/40 p-4 transition-colors hover:bg-white/5 ${
                  mine ? 'border-blood-bright' : 'border-blood/30'
                }`}
              >
                <span className="text-2xl font-semibold text-bone/50 w-7 text-center">{i + 1}</span>
                <span className="w-1.5 self-stretch rounded-full" style={{ backgroundColor: accent }} />
                {sigil && <Sigil src={sigil} />}
                <div className="flex-1">
                  <p className="text-lg font-semibold" style={{ color: accent }}>
                    House of {h.name}
                    {mine && <span className="ml-2 text-xs text-bone/50 italic">your house</span>}
                  </p>
                </div>
                <span className="text-2xl font-bold text-bone">{formatHF(h.favor)}</span>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
};
