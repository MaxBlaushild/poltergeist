import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Product } from '../api/types';
import { getSessionId } from '../lib/session';
import { Blobs } from '../components/Decor';

const CARD_ACCENTS = [
  { chip: 'bg-reef-glow/25', icon: 'text-reef-teal', hoverChip: 'group-hover:bg-reef-coral/20', hoverIcon: 'group-hover:text-reef-coral' },
  { chip: 'bg-reef-coral/20', icon: 'text-reef-coral', hoverChip: 'group-hover:bg-reef-sky/20', hoverIcon: 'group-hover:text-reef-sky' },
  { chip: 'bg-reef-sky/20', icon: 'text-reef-sky', hoverChip: 'group-hover:bg-reef-glow/25', hoverIcon: 'group-hover:text-reef-teal' },
];

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    reefApi
      .listProducts()
      .then(setProducts)
      .catch((e) => setError(String(e)));
  }, []);

  const hero = products.find((p) => p.slug === 'magnetic-frag-rack');
  const rest = products.filter((p) => p.slug !== 'magnetic-frag-rack');

  return (
    <div className="space-y-16">
      <section className="relative -mt-10 overflow-hidden px-6 pb-6 pt-16 sm:px-12">
        <Blobs />
        <div className="relative max-w-xl">
          <span className="pill border-2 border-reef-ink bg-reef-glow text-reef-ink">
            Printed to order · fit to your tank
          </span>
          <h1 className="font-display mt-5 text-5xl font-bold leading-[1.05] text-reef-ink sm:text-6xl">
            Made-to-order
            <br />
            <span className="relative inline-block">
              reef hardware
              <svg
                aria-hidden="true"
                viewBox="0 0 300 20"
                className="absolute -bottom-2 left-0 h-4 w-full text-reef-coral"
              >
                <path d="M2 12 C 60 2, 120 18, 180 8 S 260 4, 298 10" fill="none" stroke="currentColor" strokeWidth="6" strokeLinecap="round" />
              </svg>
            </span>
          </h1>
          <p className="mb-8 mt-6 max-w-md text-lg text-reef-ink/70">
            Fit to your tank, not the other way around. Configure a magnetic frag rack sized to
            your glass and rim, or grab a fixed part from the catalog.
          </p>
          {hero && (
            <Link
              to={`/configure/${hero.slug}`}
              onClick={() => reefApi.recordEvent('configurator_opened', { sessionId: getSessionId(), productSlug: hero.slug })}
              className="btn-primary text-lg"
            >
              Configure your frag rack
              <span aria-hidden="true">→</span>
            </Link>
          )}
        </div>
      </section>

      {error && <p className="text-red-600">{error}</p>}

      <section>
        <h2 className="font-display mb-5 text-2xl font-bold text-reef-ink">Catalog</h2>
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3">
          {[...(hero ? [hero] : []), ...rest].map((p, i) => (
            <ProductCard key={p.slug} product={p} accent={CARD_ACCENTS[i % CARD_ACCENTS.length]} />
          ))}
        </div>
      </section>
    </div>
  );
}

function ProductCard({ product, accent }: { product: Product; accent: (typeof CARD_ACCENTS)[number] }) {
  const href = product.kind === 'configurable' ? `/configure/${product.slug}` : `/products/${product.slug}`;
  const priceLabel =
    product.kind === 'fixed' && product.variants && product.variants.length > 0
      ? `from $${(Math.min(...product.variants.map((v) => v.priceCents)) / 100).toFixed(2)}`
      : 'built to order';
  return (
    <Link to={href} className="card card-hover group block p-5">
      <div
        className={`mb-3 flex h-10 w-10 items-center justify-center rounded-full transition-colors ${accent.chip} ${accent.icon} ${accent.hoverChip} ${accent.hoverIcon}`}
      >
        <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
          <path
            d="M3 12c2.5 2.5 5-2.5 7.5 0S15 9.5 17.5 12 21 9.5 21 9.5"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
          />
        </svg>
      </div>
      <h3 className="font-display font-bold text-reef-ink">{product.name}</h3>
      <p className="mt-1 line-clamp-3 text-sm text-reef-ink/60">{product.description}</p>
      <p className="pill mt-3 border-2 border-reef-ink/15 bg-white text-reef-ink">{priceLabel}</p>
    </Link>
  );
}
