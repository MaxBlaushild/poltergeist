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

  if (order === undefined) return <p>Loading…</p>;
  if (order === null) return <p>We couldn't find that order.</p>;

  return (
    <div className="max-w-lg space-y-4">
      <h1 className="text-2xl font-bold">Order {order.orderToken}</h1>
      <p className="rounded bg-reef-teal/10 border border-reef-teal/20 px-3 py-2 inline-block">
        Status: <span className="font-medium">{statusLabel(order.status)}</span>
      </p>

      <ul className="divide-y divide-reef-teal/10">
        {order.items.map((item) => (
          <li key={item.id} className="py-2 flex justify-between text-sm">
            <span>
              {item.variantKey ? `${item.variantKey} ` : ''}× {item.quantity}
            </span>
            <span>${((item.unitPriceCents * item.quantity) / 100).toFixed(2)}</span>
          </li>
        ))}
      </ul>

      <div className="text-sm space-y-1 border-t border-reef-teal/20 pt-3">
        <div className="flex justify-between">
          <span>Subtotal</span>
          <span>${(order.subtotalCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span>Shipping</span>
          <span>${(order.shippingCents / 100).toFixed(2)}</span>
        </div>
        <div className="flex justify-between font-semibold">
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
