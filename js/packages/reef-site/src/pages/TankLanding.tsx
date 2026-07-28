import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { TankProfile } from '../api/types';
import { toSlug } from '../lib/slug';
import { paramsToSearch } from '../lib/paramsUrl';

// R-8.3: one programmatic landing page per verified tank profile,
// pre-filling the frag rack configurator with that tank's dimensions. This
// is the primary organic-search surface for the vertical.
//
// Note: this is client-rendered like the rest of the app (see
// go/reef-site/INVENTORY.md / this package's README for why — the repo has
// no server-side-rendering precedent anywhere, and standing one up was out
// of scope here) — R-8.3 explicitly wants these pages server-rendered for
// indexability, which this does not yet deliver. Flagged, not hidden.
export default function TankLanding() {
  const { manufacturer, model } = useParams<{ manufacturer: string; model: string }>();
  const [tank, setTank] = useState<TankProfile | null | undefined>(undefined);

  useEffect(() => {
    reefApi.listTanks().then((tanks) => {
      const found = tanks.find(
        (t) => toSlug(t.manufacturer) === manufacturer && toSlug(t.model) === model,
      );
      setTank(found ?? null);
      if (found) {
        document.title = `Frag rack for the ${found.manufacturer} ${found.model} — reef`;
      }
    });
  }, [manufacturer, model]);

  if (tank === undefined) return <p className="text-reef-ink/60">Loading…</p>;
  if (tank === null) {
    return (
      <div className="card mx-auto max-w-md space-y-3 p-8 text-center">
        <p className="text-reef-ink/70">We don't have a verified profile for that tank yet.</p>
        <Link to="/configure/magnetic-frag-rack" className="btn-secondary">
          Configure by hand instead
        </Link>
      </div>
    );
  }

  const prefill = paramsToSearch({
    tankProfileId: tank.id,
    glassThicknessMm: tank.glassThicknessMm,
  });

  return (
    <div className="card space-y-4 p-8">
      <h1 className="font-display text-2xl font-bold text-reef-ink">
        Frag rack for the {tank.manufacturer} {tank.model}
      </h1>
      <p className="text-reef-ink/70">
        Pre-filled with the {tank.manufacturer} {tank.model}'s verified glass thickness (
        {tank.glassThicknessMm}mm). Fine-tune width, tiers, and hole count before ordering.
      </p>
      <Link to={`/configure/magnetic-frag-rack?${prefill}`} className="btn-primary">
        Start configuring
        <span aria-hidden="true">→</span>
      </Link>
      {tank.sourceUrl && (
        <p className="text-xs text-reef-ink/40">
          Dimensions sourced from{' '}
          <a href={tank.sourceUrl} target="_blank" rel="noreferrer" className="underline hover:text-reef-teal">
            manufacturer specs
          </a>
          .
        </p>
      )}
    </div>
  );
}
