import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Product } from '../api/types';
import { getSessionId } from '../lib/session';

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
    <div className="space-y-12">
      <section className="rounded-lg bg-reef-deep text-reef-sand p-8">
        <h1 className="text-3xl font-bold mb-2">Made-to-order reef hardware</h1>
        <p className="max-w-xl text-reef-sand/80 mb-6">
          Fit to your tank, not the other way around. Configure a magnetic frag rack sized to your
          glass and rim, or grab a fixed part from the catalog.
        </p>
        {hero && (
          <Link
            to={`/configure/${hero.slug}`}
            onClick={() => reefApi.recordEvent('configurator_opened', { sessionId: getSessionId(), productSlug: hero.slug })}
            className="inline-block rounded bg-reef-coral px-5 py-3 font-semibold text-reef-ink hover:opacity-90"
          >
            Configure your frag rack
          </Link>
        )}
      </section>

      {error && <p className="text-red-600">{error}</p>}

      <section>
        <h2 className="text-xl font-semibold mb-4">Catalog</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {[...(hero ? [hero] : []), ...rest].map((p) => (
            <ProductCard key={p.slug} product={p} />
          ))}
        </div>
      </section>
    </div>
  );
}

function ProductCard({ product }: { product: Product }) {
  const href = product.kind === 'configurable' ? `/configure/${product.slug}` : `/products/${product.slug}`;
  const priceLabel =
    product.kind === 'fixed' && product.variants && product.variants.length > 0
      ? `from $${(Math.min(...product.variants.map((v) => v.priceCents)) / 100).toFixed(2)}`
      : 'built to order';
  return (
    <Link
      to={href}
      className="block rounded-lg border border-reef-teal/20 bg-white p-4 hover:shadow-md transition-shadow"
    >
      <h3 className="font-semibold">{product.name}</h3>
      <p className="text-sm text-reef-ink/70 mt-1 line-clamp-3">{product.description}</p>
      <p className="text-sm mt-2 text-reef-teal">{priceLabel}</p>
    </Link>
  );
}
