import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Product, ProductVariant } from '../api/types';
import { useCart } from '../hooks/useCart';
import { getSessionId } from '../lib/session';

export default function ProductDetail() {
  const { slug } = useParams<{ slug: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null);
  const [added, setAdded] = useState(false);
  const { addItem } = useCart();

  useEffect(() => {
    if (!slug) return;
    reefApi.getProduct(slug).then((p) => {
      setProduct(p);
      setSelectedVariant(p.variants?.[0] ?? null);
    });
  }, [slug]);

  if (!product) return <p>Loading…</p>;

  return (
    <div className="max-w-lg space-y-4">
      <h1 className="text-2xl font-bold">{product.name}</h1>
      <p className="text-reef-ink/80">{product.description}</p>
      <p className="text-sm text-reef-ink/60">Material: {product.material}</p>

      {product.variants && product.variants.length > 0 && (
        <div>
          <label className="block text-sm font-medium mb-1">Size</label>
          <select
            className="border border-reef-teal/30 rounded px-3 py-2"
            value={selectedVariant?.variantKey ?? ''}
            onChange={(e) => setSelectedVariant(product.variants!.find((v) => v.variantKey === e.target.value) ?? null)}
          >
            {product.variants.map((v) => (
              <option key={v.variantKey} value={v.variantKey}>
                {v.label} — ${(v.priceCents / 100).toFixed(2)}
              </option>
            ))}
          </select>
        </div>
      )}

      <button
        className="rounded bg-reef-coral px-5 py-3 font-semibold text-reef-ink hover:opacity-90 disabled:opacity-50"
        disabled={!selectedVariant}
        onClick={() => {
          if (!selectedVariant) return;
          addItem({ productSlug: product.slug, variantKey: selectedVariant.variantKey, quantity: 1 });
          reefApi.recordEvent('add_to_cart', { sessionId: getSessionId(), productSlug: product.slug });
          setAdded(true);
        }}
      >
        Add to cart
      </button>
      {added && <p className="text-sm text-reef-teal">Added to cart.</p>}
    </div>
  );
}
