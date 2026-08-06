import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Configuration, ParameterSchema } from '../api/types';
import StlViewer from '../components/StlViewer';

function capitalize(s: string): string {
  return s.length ? s[0].toUpperCase() + s.slice(1) : s;
}

function formatValue(value: unknown, unit?: string): string {
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (typeof value === 'string') return unit ? `${capitalize(value)} ${unit}` : capitalize(value);
  if (typeof value === 'number') return unit ? `${value} ${unit}` : String(value);
  return String(value);
}

// What an order item's "view part" link lands on — read-only, no
// parameters to edit, just the exact STL that was sliced for this order
// (R-4.6's viewer, reused rather than duplicated).
export default function ConfigurationPreview() {
  const { id } = useParams<{ id: string }>();
  const [configuration, setConfiguration] = useState<Configuration | null | undefined>(undefined);
  const [schema, setSchema] = useState<ParameterSchema | null>(null);

  useEffect(() => {
    if (!id) return;
    reefApi
      .getConfiguration(id)
      .then(setConfiguration)
      .catch(() => setConfiguration(null));
  }, [id]);

  useEffect(() => {
    if (!configuration?.productSlug) return;
    reefApi
      .getProductSchema(configuration.productSlug)
      .then(setSchema)
      .catch(() => setSchema(null));
  }, [configuration?.productSlug]);

  if (configuration === undefined) return <p className="text-reef-ink/60">Loading…</p>;
  if (configuration === null) return <p className="text-reef-ink/60">We couldn't find that part.</p>;

  // Schema-driven, human-friendly labels/units (same x-label/x-unit
  // metadata SchemaForm.tsx renders the configurator from) rather than
  // dumping raw params keys like "widthMm"/"tierCount" — fields with no
  // value (e.g. tankProfileId left unset) are skipped entirely.
  const entries = Object.entries(configuration.params).filter(([, value]) => value !== null && value !== undefined);

  return (
    <div className="mx-auto max-w-lg space-y-5">
      <Link to="/" className="text-sm text-reef-ink/60 hover:text-reef-coral">
        ← Back
      </Link>
      <h1 className="font-display text-2xl font-bold text-reef-ink">Your part</h1>
      <StlViewer url={configuration.previewUrl ?? null} pending={false} plateFits />
      <dl className="card grid grid-cols-2 gap-x-4 gap-y-2 p-5 text-sm">
        {entries.map(([key, value]) => {
          const prop = schema?.properties[key];
          const label = prop?.['x-label'] ?? capitalize(key);
          return (
            <div key={key} className="contents">
              <dt className="text-reef-ink/60">{label}</dt>
              <dd className="text-right font-medium text-reef-ink">{formatValue(value, prop?.['x-unit'])}</dd>
            </div>
          );
        })}
      </dl>
    </div>
  );
}
