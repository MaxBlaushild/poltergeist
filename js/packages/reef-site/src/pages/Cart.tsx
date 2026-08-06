import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { CartResponse } from '../api/types';
import { useCart } from '../hooks/useCart';
import { useCustomerAuth } from '../hooks/useCustomerAuth';
import { getSessionId } from '../lib/session';

export default function Cart() {
  const { items, removeItem, setQuantity } = useCart();
  const { auth } = useCustomerAuth();
  const [cart, setCart] = useState<CartResponse | null>(null);
  const [email, setEmail] = useState(() => auth?.user.email ?? '');
  const [checkoutError, setCheckoutError] = useState<string | null>(null);
  const [checkingOut, setCheckingOut] = useState(false);

  useEffect(() => {
    // When items is empty the component returns its own "cart is empty"
    // view below before ever reading `cart`, so there's nothing to
    // synchronize here.
    if (items.length === 0) return;
    reefApi.cart(items).then(setCart);
  }, [items]);

  // Keeps `email` correct if login happens after this page is already
  // mounted (e.g. logging in from another tab) — the state above only
  // captures auth at first render.
  useEffect(() => {
    if (auth?.user.email) setEmail(auth.user.email);
  }, [auth]);

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
      <div className="card mx-auto max-w-md space-y-3 p-8 text-center">
        <p className="text-reef-ink/70">Your cart is empty.</p>
        <Link to="/" className="btn-secondary">
          Back to the catalog
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="font-display text-2xl font-bold text-reef-ink">Cart</h1>

      {!cart && <p className="text-reef-ink/60">Loading…</p>}

      {cart && (
        <>
          <ul className="card divide-y divide-reef-teal/10 px-5">
            {cart.items.map((item, i) => (
              <li key={i} className="flex items-center justify-between py-4">
                <div>
                  <p className="font-medium text-reef-ink">
                    {item.productName}
                    {item.variantLabel ? ` — ${item.variantLabel}` : ''}
                  </p>
                  <p className="text-sm text-reef-ink/60">${(item.unitPriceCents / 100).toFixed(2)} each</p>
                  <div className="mt-2 flex items-center gap-2 text-sm">
                    <button
                      className="btn-ghost"
                      aria-label="Decrease quantity"
                      onClick={() => setQuantity(items[i], Math.max(0, item.quantity - 1))}
                    >
                      −
                    </button>
                    <span className="w-4 text-center tabular-nums">{item.quantity}</span>
                    <button
                      className="btn-ghost"
                      aria-label="Increase quantity"
                      onClick={() => setQuantity(items[i], item.quantity + 1)}
                    >
                      +
                    </button>
                    <button
                      className="ml-2 text-xs font-medium text-red-500 underline underline-offset-2 hover:text-red-600"
                      onClick={() => removeItem(items[i])}
                    >
                      Remove
                    </button>
                  </div>
                </div>
                <p className="font-medium text-reef-ink">${(item.lineTotalCents / 100).toFixed(2)}</p>
              </li>
            ))}
          </ul>

          {cart.remainingToFreeShippingCents > 0 && (
            <p className="pill w-full justify-center bg-reef-glow/15 py-2 text-sm text-reef-ink ring-1 ring-inset ring-reef-glow/30">
              Add ${(cart.remainingToFreeShippingCents / 100).toFixed(2)} more for free shipping.
            </p>
          )}

          {cart.crossSell && cart.crossSell.length > 0 && (
            <div>
              <p className="mb-2 text-sm font-medium text-reef-ink">You might also need:</p>
              <div className="flex flex-wrap gap-3">
                {cart.crossSell.map((p) => (
                  <Link key={p.slug} to={`/products/${p.slug}`} className="btn-secondary py-2 text-sm">
                    {p.name}
                  </Link>
                ))}
              </div>
            </div>
          )}

          <div className="card space-y-1 p-5 text-sm">
            <div className="flex justify-between">
              <span className="text-reef-ink/70">Subtotal</span>
              <span>${(cart.subtotalCents / 100).toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-reef-ink/70">Shipping</span>
              <span>{cart.shippingCents === 0 ? 'Free' : `$${(cart.shippingCents / 100).toFixed(2)}`}</span>
            </div>
            <div className="flex justify-between border-t border-reef-teal/10 pt-2 text-base font-semibold text-reef-ink">
              <span>Total</span>
              <span>${(cart.totalCents / 100).toFixed(2)}</span>
            </div>
          </div>

          {auth?.user.email ? (
            <p className="text-sm text-reef-ink/70">
              Checking out as <span className="font-medium text-reef-ink">{auth.user.email}</span>
            </p>
          ) : (
            <div>
              <label className="mb-1 block text-sm font-medium text-reef-ink">Email</label>
              <input
                type="email"
                className="input-field"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
              />
            </div>
          )}

          {checkoutError && <p className="text-sm text-red-600">{checkoutError}</p>}

          <button onClick={handleCheckout} disabled={checkingOut} className="btn-primary w-full">
            {checkingOut ? 'Redirecting to checkout…' : 'Checkout'}
          </button>
        </>
      )}
    </div>
  );
}
