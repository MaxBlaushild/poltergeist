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

  if (!product) return <p className="text-reef-ink/60">Loading…</p>;

  return (
    <div className="card max-w-lg space-y-4 p-6">
      <h1 className="font-display text-2xl font-bold text-reef-ink">{product.name}</h1>
      <p className="text-reef-ink/80">{product.description}</p>
      <p className="pill bg-reef-teal/10 text-reef-teal">Material: {product.material}</p>

      {product.variants && product.variants.length > 0 && (
        <div>
          <label className="mb-1 block text-sm font-medium">Size</label>
          <select
            className="input-field"
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
        className="btn-primary w-full"
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
      {added && <p className="text-sm font-medium text-reef-teal">✓ Added to cart.</p>}
    </div>
  );
}
