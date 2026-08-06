import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { bgiApi } from '../api/client';
import type { GameDetail as GameDetailData } from '../api/types';

// A landing page for a game before the customer enters the configurator —
// mostly useful once R-8.3's programmatic per-game/expansion pages exist
// (deferred, see PLATFORM_FINDINGS.md); for v1 it's the honest-data landing
// spot linked from anywhere a game is mentioned outside the configurator.
export default function GameDetail() {
  const { slug } = useParams<{ slug: string }>();
  const [game, setGame] = useState<GameDetailData | null>(null);

  useEffect(() => {
    if (!slug) return;
    bgiApi.getGame(slug).then(setGame);
  }, [slug]);

  if (!game) return <p className="text-bgi-ink/60">Loading…</p>;

  const anyUnverified =
    game.boxProfiles.some((b) => !b.verified) ||
    game.sleeveProfiles.some((s) => !s.verified) ||
    game.manifest.some((m) => !m.verified);

  return (
    <div className="card mx-auto max-w-lg space-y-4 p-6">
      <h1 className="font-display text-2xl font-bold text-bgi-ink">{game.name}</h1>
      <p className="text-bgi-ink/70">
        {game.publisher}
        {game.yearPublished ? ` · ${game.yearPublished}` : ''}
      </p>

      {anyUnverified && (
        <p className="banner-caution">
          Box and sleeve dimensions for this game are sourced from published/community references and pending
          physical verification.{' '}
          <Link to={`/games/${slug}/compatibility`} className="underline decoration-bgi-coral/50 underline-offset-2">
            Read more
          </Link>
          .
        </p>
      )}

      {game.productSlug ? (
        <Link to={`/configure/${game.slug}`} className="btn-primary w-full">
          Configure your tray set
        </Link>
      ) : (
        <p className="text-sm text-bgi-ink/60">No configurable product yet for this game.</p>
      )}

      <p className="text-xs text-bgi-ink/50">
        Unaffiliated with and unendorsed by {game.publisher || 'the publisher'}. See{' '}
        <Link to={`/games/${slug}/compatibility`} className="underline decoration-bgi-teal/40 underline-offset-2">
          compatibility notes
        </Link>
        .
      </p>
    </div>
  );
}
