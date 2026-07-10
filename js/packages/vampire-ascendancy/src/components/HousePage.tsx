import { useEffect, useState } from 'react';
import { Navigate, useNavigate, useParams } from 'react-router-dom';
import { getHouseOverview, getToken } from '../api';
import type { HouseOverview } from '../types';
import { accentFor, houseInfoFor, formatHF, taglineFor, houseLabel } from '../theme';

export const HousePage = () => {
  const token = getToken();
  const { houseId } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState<HouseOverview | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    if (!token || !houseId) return;
    let cancelled = false;
    const load = () =>
      getHouseOverview(token, houseId)
        .then((d) => !cancelled && setData(d))
        .catch(() => !cancelled && setError(true));
    load();
    const id = setInterval(load, 5000); // keep the favor log live
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [token, houseId]);

  if (!token) return <Navigate to="/login" replace />;

  if (error) {
    return (
      <Centered>
        <p className="text-bone/80">That house could not be found.</p>
        <button onClick={() => navigate(-1)} className="mt-4 text-gold uppercase tracking-[0.2em] text-sm">
          ← Back
        </button>
      </Centered>
    );
  }
  if (!data) return <Centered>Consulting the archives…</Centered>;

  const { house, members, log } = data;
  const accent = accentFor(house.name);
  const info = houseInfoFor(house.name);

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 pt-3 pb-12">
      <button
        onClick={() => navigate(-1)}
        className="text-bone/60 hover:text-bone uppercase tracking-[0.2em] text-xs mb-4"
      >
        ← Back
      </button>

      <header className="text-center mb-6">
        {info.sigil && <HouseSigil src={info.sigil} />}
        <p className="text-xs uppercase tracking-[0.4em] text-gold">The Crimson Toast</p>
        <h1 className="mt-2 font-display text-4xl font-bold leading-tight" style={{ color: accent }}>
          {houseLabel(house.name)}
        </h1>
        {taglineFor(house.name) && (
          <p className="mt-3 font-bold" style={{ color: accent }}>
            {taglineFor(house.name)}
          </p>
        )}
        {info.blurb && <p className="mt-2 text-bone/85 leading-relaxed">{info.blurb}</p>}
        <div className="mt-4 inline-flex items-baseline gap-2">
          <span className="text-bone/60 uppercase tracking-[0.2em] text-xs">House Favor</span>
          <span className="text-3xl font-bold text-bone">
            {formatHF(house.favor)}
            {house.itemFavor ? <span className="ml-1 text-xl text-gold">(+{formatHF(house.itemFavor)})</span> : null}
          </span>
        </div>
      </header>

      <Section title="Favor Ledger">
        {log.length === 0 ? (
          <p className="text-bone/50">No favor recorded yet.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {log.map((e) => (
              <div
                key={e.id}
                className="flex items-start gap-3 rounded-lg border border-blood/30 bg-black/40 px-4 py-3"
              >
                <span
                  className={`font-bold w-10 shrink-0 text-right ${
                    e.delta >= 0 ? 'text-green-400' : 'text-blood-bright'
                  }`}
                >
                  {e.delta >= 0 ? `+${formatHF(e.delta)}` : formatHF(e.delta)}
                </span>
                <div className="flex-1">
                  <p className="text-bone">{e.reason || (e.source === 'quiz' ? 'Quiz result' : 'Adjustment')}</p>
                  <p className="text-bone/40 text-xs">{formatWhen(e.createdAt)}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>

      <Section title="Members">
        <div className="flex flex-col gap-1.5">
          {members.map((m) => (
            <div key={m.id} className="rounded-lg border border-blood/30 bg-black/40 px-4 py-2">
              <p className="text-bone">
                {m.name}
                <span className="text-bone/40 mx-2">•</span>
                <span className="text-bone/55 italic">{m.title}</span>
              </p>
            </div>
          ))}
        </div>
      </Section>
    </div>
  );
};

// Hides itself if the sigil image hasn't been added yet. The source art is black
// line-work on white, so we invert it to read as a light emblem on the dark page.
const HouseSigil = ({ src }: { src: string; accent?: string }) => {
  const [failed, setFailed] = useState(false);
  if (failed) return null;
  return (
    <img
      src={src}
      alt=""
      onError={() => setFailed(true)}
      className="w-24 h-24 mx-auto mb-3 object-contain"
      // invert -> light emblem on black; screen blend drops the (now black) bg square
      style={{ filter: 'invert(1)', mixBlendMode: 'screen' }}
    />
  );
};

const formatWhen = (iso: string) => {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
};

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="mt-6">
    <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-2">{title}</h2>
    {children}
  </section>
);

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex flex-col items-center justify-center px-6 text-center text-bone/80">
    {children}
  </div>
);
