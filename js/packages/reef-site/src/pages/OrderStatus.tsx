import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { Order } from '../api/types';

// R-8.2: /orders/[token] — order status, no login.
export default function OrderStatus() {
  const { token } = useParams<{ token: string }>();
  const [order, setOrder] = useState<Order | null | undefined>(undefined);

  useEffect(() => {
    if (!token) return;
    reefApi
      .getOrder(token)
      .then(setOrder)
      .catch(() => setOrder(null));
  }, [token]);

  if (order === undefined) return <p className="text-reef-ink/60">Loading…</p>;
  if (order === null) return <p className="text-reef-ink/60">We couldn't find that order.</p>;

  return (
    <div className="mx-auto max-w-lg space-y-5">
      <h1 className="font-display text-2xl font-bold text-reef-ink">Order {order.orderToken}</h1>
      <p className="pill bg-reef-teal/10 text-reef-teal">
        Status: <span className="ml-1 font-semibold">{statusLabel(order.status)}</span>
      </p>

      <ul className="card divide-y divide-reef-teal/10 px-5">
        {order.items.map((item) => (
          <li key={item.id} className="flex justify-between py-3 text-sm">
            <span className="text-reef-ink/80">
              {item.variantKey ? `${item.variantKey} ` : ''}× {item.quantity}
            </span>
            <span className="font-medium text-reef-ink">${((item.unitPriceCents * item.quantity) / 100).toFixed(2)}</span>
          </li>
        ))}
      </ul>

      <div className="card space-y-1 p-5 text-sm">
        <div className="flex justify-between">
          <span className="text-reef-ink/70">Subtotal</span>
          <span>${(order.subtotalCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-reef-ink/70">Shipping</span>
          <span>${(order.shippingCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between border-t border-reef-teal/10 pt-2 font-semibold text-reef-ink">
          <span>Total</span>
          <span>${(order.totalCents / 100).toFixed(2)}</span>
        </div>
      </div>
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
