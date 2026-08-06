import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { bgiApi } from '../api/client';
import type { Order } from '../api/types';
import { useCart } from '../hooks/useCart';

// /orders/[token] — order status, no login (R-1.3).
export default function OrderStatus() {
  const { token } = useParams<{ token: string }>();
  const [order, setOrder] = useState<Order | null | undefined>(undefined);
  const { clear } = useCart();

  useEffect(() => {
    if (!token) return;
    bgiApi
      .getOrder(token)
      .then((data) => {
        setOrder(data);
        // Same reasoning as reef-site's own: this page is only reachable
        // via Stripe's real success redirect or a past-order revisit, so
        // the cart that led here is stale.
        clear();
      })
      .catch(() => setOrder(null));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  if (order === undefined) return <p className="text-bgi-ink/60">Loading…</p>;
  if (order === null) return <p className="text-bgi-ink/60">We couldn't find that order.</p>;

  return (
    <div className="mx-auto max-w-lg space-y-5">
      <h1 className="font-display text-2xl font-bold text-bgi-ink">Order {order.orderToken}</h1>
      <p className="pill bg-bgi-teal/10 text-bgi-teal">
        Status: <span className="ml-1 font-semibold">{statusLabel(order.status)}</span>
      </p>

      <ul className="card divide-y divide-bgi-teal/10 px-5">
        {order.items.map((item) => (
          <li key={item.id} className="flex justify-between py-3 text-sm">
            <span className="text-bgi-ink/80">
              {item.productName || 'Tray set'} × {item.quantity}
            </span>
            <span className="font-medium text-bgi-ink">${((item.unitPriceCents * item.quantity) / 100).toFixed(2)}</span>
          </li>
        ))}
      </ul>

      <div className="card space-y-1 p-5 text-sm">
        <div className="flex justify-between">
          <span className="text-bgi-ink/70">Subtotal</span>
          <span>${(order.subtotalCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-bgi-ink/70">Shipping</span>
          <span>${(order.shippingCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between border-t border-bgi-teal/10 pt-2 font-semibold text-bgi-ink">
          <span>Total</span>
          <span>${(order.totalCents / 100).toFixed(2)}</span>
        </div>
      </div>

      {order.status === 'pending_payment' && (
        <p className="text-xs text-bgi-ink/50">
          Once payment clears, this order is generated as N supportless, single-plate trays and queued for printing —
          expect several days of fulfillment lead time.
        </p>
      )}
    </div>
  );
}

function statusLabel(status: Order['status']): string {
  switch (status) {
    case 'pending_payment':
      return 'Awaiting payment';
    case 'paid':
      return 'Paid — queued for printing';
    case 'fulfilled':
      return 'Shipped';
    case 'cancelled':
      return 'Cancelled';
    default:
      return status;
  }
}
