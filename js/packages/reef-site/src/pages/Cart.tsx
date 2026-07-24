import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { CartResponse } from '../api/types';
import { useCart } from '../hooks/useCart';
import { getSessionId } from '../lib/session';

export default function Cart() {
  const { items, removeItem, setQuantity } = useCart();
  const [cart, setCart] = useState<CartResponse | null>(null);
  const [email, setEmail] = useState('');
  const [checkoutError, setCheckoutError] = useState<string | null>(null);
  const [checkingOut, setCheckingOut] = useState(false);

  useEffect(() => {
    // When items is empty the component returns its own "cart is empty"
    // view below before ever reading `cart`, so there's nothing to
    // synchronize here.
    if (items.length === 0) return;
    reefApi.cart(items).then(setCart);
  }, [items]);

  const handleCheckout = async () => {
    if (!email) {
      setCheckoutError('Enter an email address to continue.');
      return;
    }
    setCheckingOut(true);
    setCheckoutError(null);
    reefApi.recordEvent('checkout_started', { sessionId: getSessionId() });
    try {
      const result = await reefApi.checkout(
        items,
        email,
        `${window.location.origin}/orders/:orderToken:`,
        `${window.location.origin}/cart`,
        getSessionId(),
      );
      window.location.href = result.checkoutUrl;
    } catch (e) {
      setCheckoutError(e instanceof Error ? e.message : 'Checkout failed');
      setCheckingOut(false);
    }
  };

  if (items.length === 0) {
    return (
      <div>
        <p>Your cart is empty.</p>
        <Link to="/" className="text-reef-teal underline">
          Back to the catalog
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-lg space-y-6">
      <h1 className="text-2xl font-bold">Cart</h1>

      {!cart && <p>Loading…</p>}

      {cart && (
        <>
          <ul className="divide-y divide-reef-teal/10">
            {cart.items.map((item, i) => (
              <li key={i} className="py-3 flex items-center justify-between">
                <div>
                  <p className="font-medium">
                    {item.productName}
                    {item.variantLabel ? ` — ${item.variantLabel}` : ''}
                  </p>
                  <p className="text-sm text-reef-ink/60">${(item.unitPriceCents / 100).toFixed(2)} each</p>
                  <div className="flex items-center gap-2 mt-1 text-sm">
                    <button
                      className="underline"
                      onClick={() => setQuantity(items[i], Math.max(0, item.quantity - 1))}
                    >
                      −
                    </button>
                    <span>{item.quantity}</span>
                    <button className="underline" onClick={() => setQuantity(items[i], item.quantity + 1)}>
                      +
                    </button>
                    <button className="ml-3 text-red-600 underline" onClick={() => removeItem(items[i])}>
                      Remove
                    </button>
                  </div>
                </div>
                <p className="font-medium">${(item.lineTotalCents / 100).toFixed(2)}</p>
              </li>
            ))}
          </ul>

          {cart.remainingToFreeShippingCents > 0 && (
            <p className="text-sm rounded bg-reef-deep/5 border border-reef-teal/20 px-3 py-2">
              Add ${(cart.remainingToFreeShippingCents / 100).toFixed(2)} more for free shipping.
            </p>
          )}

          {cart.crossSell && cart.crossSell.length > 0 && (
            <div>
              <p className="text-sm font-medium mb-2">You might also need:</p>
              <div className="flex gap-3">
                {cart.crossSell.map((p) => (
                  <Link
                    key={p.slug}
                    to={`/products/${p.slug}`}
                    className="text-sm rounded border border-reef-teal/20 px-3 py-2 hover:bg-reef-teal/5"
                  >
                    {p.name}
                  </Link>
                ))}
              </div>
            </div>
          )}

          <div className="text-sm space-y-1 border-t border-reef-teal/20 pt-4">
            <div className="flex justify-between">
              <span>Subtotal</span>
              <span>${(cart.subtotalCents / 100).toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span>Shipping</span>
              <span>{cart.shippingCents === 0 ? 'Free' : `$${(cart.shippingCents / 100).toFixed(2)}`}</span>
            </div>
            <div className="flex justify-between font-semibold text-base">
              <span>Total</span>
              <span>${(cart.totalCents / 100).toFixed(2)}</span>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              className="w-full border border-reef-teal/30 rounded px-3 py-2"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
            />
          </div>

          {checkoutError && <p className="text-sm text-red-600">{checkoutError}</p>}

          <button
            onClick={handleCheckout}
            disabled={checkingOut}
            className="w-full rounded bg-reef-coral px-5 py-3 font-semibold text-reef-ink hover:opacity-90 disabled:opacity-50"
          >
            {checkingOut ? 'Redirecting to checkout…' : 'Checkout'}
          </button>
        </>
      )}
    </div>
  );
}
