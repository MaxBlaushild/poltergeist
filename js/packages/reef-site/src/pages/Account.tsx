import { useEffect, useState } from 'react';
import { Link, Navigate } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { MyOrder } from '../api/types';
import { useCustomerAuth } from '../hooks/useCustomerAuth';

function usd(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function statusLabel(status: string): string {
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

// R-8.2 extended: order history for logged-in customers, alongside (not
// replacing) the anonymous /orders/[token] lookup every order confirmation
// still links to.
export default function Account() {
  const { auth, logout } = useCustomerAuth();
  const [orders, setOrders] = useState<MyOrder[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!auth) return;
    reefApi
      .myOrders()
      .then(setOrders)
      .catch(() => setError('Failed to load your orders'));
  }, [auth]);

  if (!auth) return <Navigate to="/login" replace />;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-display text-2xl font-bold text-reef-ink">My orders</h1>
          <p className="text-sm text-reef-ink/60">{auth.user.name}</p>
        </div>
        <button onClick={logout} className="btn-secondary py-2 text-sm">
          Log out
        </button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
      {orders === null && !error && <p className="text-reef-ink/60">Loading…</p>}
      {orders && orders.length === 0 && <p className="text-reef-ink/60">No orders yet.</p>}

      <div className="space-y-4">
        {orders?.map((order) => (
          <div key={order.id} className="card space-y-3 p-5 text-sm">
            <div className="flex items-center justify-between">
              <Link
                to={`/orders/${order.orderToken}`}
                className="font-semibold text-reef-ink underline decoration-reef-teal/40 underline-offset-2 hover:text-reef-coral"
              >
                Order {order.orderToken}
              </Link>
              <span className="pill bg-reef-teal/10 text-reef-teal">{statusLabel(order.status)}</span>
            </div>
            <ul className="divide-y divide-reef-teal/10 border-t border-reef-teal/10">
              {order.items.map((item) => (
                <li key={item.id} className="flex items-center justify-between py-2">
                  <span>
                    {item.productName || 'Item'}
                    {item.variantKey ? ` (${item.variantKey})` : ''} × {item.quantity}
                  </span>
                  <span className="text-reef-ink/70">{usd(item.unitPriceCents * item.quantity)}</span>
                </li>
              ))}
            </ul>
            <p className="text-right font-semibold text-reef-ink">Total: {usd(order.totalCents)}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
