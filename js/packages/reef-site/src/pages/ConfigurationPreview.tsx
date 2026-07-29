import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Configuration } from '../api/types';
import StlViewer from '../components/StlViewer';

// What an order item's "view part" link lands on — read-only, no
// parameters to edit, just the exact STL that was sliced for this order
// (R-4.6's viewer, reused rather than duplicated).
export default function ConfigurationPreview() {
  const { id } = useParams<{ id: string }>();
  const [configuration, setConfiguration] = useState<Configuration | null | undefined>(undefined);

  useEffect(() => {
    if (!id) return;
    reefApi
      .getConfiguration(id)
      .then(setConfiguration)
      .catch(() => setConfiguration(null));
  }, [id]);

  if (configuration === undefined) return <p className="text-reef-ink/60">Loading…</p>;
  if (configuration === null) return <p className="text-reef-ink/60">We couldn't find that part.</p>;

  return (
    <div className="mx-auto max-w-lg space-y-5">
      <Link to="/" className="text-sm text-reef-ink/60 hover:text-reef-coral">
        ← Back
      </Link>
      <h1 className="font-display text-2xl font-bold text-reef-ink">Your part</h1>
      <StlViewer url={configuration.previewUrl ?? null} pending={false} plateFits />
      <dl className="card grid grid-cols-2 gap-x-4 gap-y-2 p-5 text-sm">
        {Object.entries(configuration.params).map(([key, value]) => (
          <div key={key} className="contents">
            <dt className="text-reef-ink/60">{key}</dt>
            <dd className="text-right font-medium text-reef-ink">{String(value)}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
