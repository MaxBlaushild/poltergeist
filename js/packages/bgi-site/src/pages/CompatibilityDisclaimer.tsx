import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { bgiApi } from '../api/client';
import type { CompatibilityInfo, Game } from '../api/types';

// /games/[slug]/compatibility (R-8.4/R-2.5). Nominative-use statement,
// no-affiliation disclaimer, no publisher logos anywhere in this codebase's
// assets — linked from every product/game page. This copy is a starting
// point, not final legal language; it should be reviewed by the site
// owner/counsel before real launch (this build's own scope explicitly
// stops short of that review — see PLATFORM_FINDINGS.md).
export default function CompatibilityDisclaimer() {
  const { slug } = useParams<{ slug: string }>();
  const [game, setGame] = useState<Game | null>(null);
  const [info, setInfo] = useState<CompatibilityInfo | null>(null);

  useEffect(() => {
    if (!slug) return;
    bgiApi.getGame(slug).then(setGame);
    bgiApi.getGameCompatibility(slug).then(setInfo);
  }, [slug]);

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="font-display text-2xl font-bold text-bgi-ink">Compatibility &amp; disclaimer</h1>

      <div className="card space-y-3 p-6 text-sm text-bgi-ink/80">
        <p>
          Products on this site are organizer trays sized to fit specific commercial board games, described using
          the game's name for the sole purpose of indicating compatibility (nominative reference). This site is{' '}
          <strong>not affiliated with, endorsed by, or sponsored by</strong> the publisher of {game?.name ?? 'any game listed here'}
          {game?.publisher ? ` (${game.publisher})` : ''} or any other game referenced on this site.
        </p>
        <p>
          No box art, logos, trademarked graphics, or fonts belonging to any publisher are reproduced on this site
          or on any printed part — only functional geometry sized to publicly known, factual component dimensions
          and counts.
        </p>
        <p>
          Box dimensions can vary between printings and regional editions. Component counts and card sizes are
          sourced from published community references and may be revised. No warranty of exact fit is made for any
          game not yet marked verified below.
        </p>
      </div>

      {game && info && (
        <div className="card space-y-4 p-6">
          <h2 className="font-display text-lg font-bold text-bgi-ink">{game.name} — data sources</h2>

          <div>
            <h3 className="mb-1 text-sm font-semibold text-bgi-ink">Box dimensions</h3>
            <ul className="space-y-1 text-sm text-bgi-ink/70">
              {info.boxProfiles.map((b) => (
                <li key={b.id}>
                  {b.label} ({b.source}): {b.interiorLengthMm}×{b.interiorWidthMm}×{b.interiorDepthMm}mm interior —{' '}
                  {b.verified ? (
                    <span className="text-bgi-teal">verified</span>
                  ) : (
                    <span className="text-bgi-coral">unverified{b.depthIsPlaceholder ? ', depth is a placeholder' : ''}</span>
                  )}
                  {b.measurementNotes && <span className="block text-xs italic">{b.measurementNotes}</span>}
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="mb-1 text-sm font-semibold text-bgi-ink">Component counts</h3>
            <ul className="space-y-1 text-sm text-bgi-ink/70">
              {info.manifest.map((m) => (
                <li key={m.id}>
                  {m.componentType}: {m.count} —{' '}
                  {m.verified ? <span className="text-bgi-teal">verified</span> : <span className="text-bgi-coral">unverified</span>}
                  {m.notes && <span className="block text-xs italic">{m.notes}</span>}
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}

      <Link to="/" className="btn-secondary">
        Back to the catalog
      </Link>
    </div>
  );
}
