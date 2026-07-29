import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { reefApi } from '../api/client';
import type { OperatorOrder } from '../api/types';
import AdminAuthGate, { isUnauthorized } from '../components/AdminAuthGate';

function usd(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function shipToLine(order: OperatorOrder): string {
  const addr = order.shippingAddress;
  if (!addr) return '(no shipping address on file)';
  return `${addr.name} — ${addr.line1}${addr.line2 ? ' ' + addr.line2 : ''}, ${addr.city}, ${addr.state} ${addr.postalCode}, ${addr.country}`;
}

function nextAction(status: string): { label: string; next: 'printed' | 'shipped' } | null {
  if (status === 'printed') return { label: 'Mark shipped', next: 'shipped' };
  if (status === 'shipped') return null;
  // "submitted" or a submission_failed:* string — printing is always the
  // next manual step regardless of exactly how it got here.
  return { label: 'Mark printed', next: 'printed' };
}

export default function PrintQueue() {
  return <AdminAuthGate>{(onAuthError) => <PrintQueueView onAuthError={onAuthError} />}</AdminAuthGate>;
}

function PrintQueueView({ onAuthError }: { onAuthError: () => void }) {
  const [orders, setOrders] = useState<OperatorOrder[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  const load = () => {
    reefApi
      .operatorOrders()
      .then(setOrders)
      .catch((err) => (isUnauthorized(err) ? onAuthError() : setError('Failed to load orders')));
  };

  useEffect(load, []);

  const handleAdvance = async (order: OperatorOrder, next: 'printed' | 'shipped') => {
    setUpdatingId(order.id);
    try {
      const updated = await reefApi.updateOrderFulfillment(order.id, next);
      setOrders((prev) => prev && prev.map((o) => (o.id === order.id ? updated : o)));
    } catch (err) {
      if (isUnauthorized(err)) onAuthError();
      else setError('Failed to update order');
    } finally {
      setUpdatingId(null);
    }
  };

  const handleSlantAction = async (order: OperatorOrder, action: 'send' | 'refresh') => {
    setUpdatingId(order.id);
    setError(null);
    try {
      const updated =
        action === 'send' ? await reefApi.fulfillOrderWithSlant(order.id) : await reefApi.refreshSlantStatus(order.id);
      setOrders((prev) => prev && prev.map((o) => (o.id === order.id ? updated : o)));
    } catch (err) {
      if (isUnauthorized(err)) onAuthError();
      else setError(err instanceof Error ? err.message : 'Slant request failed');
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold text-reef-lagoon">Print queue</h1>
        <Link to="/operator" className="text-sm font-medium text-reef-teal underline underline-offset-2">
          ← Metrics
        </Link>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
      {orders === null && !error && <p className="text-reef-ink/60">Loading…</p>}
      {orders && orders.length === 0 && <p className="text-reef-ink/60">No paid orders yet.</p>}

      <div className="space-y-4">
        {orders?.map((order) => {
          const action = nextAction(order.fulfillmentStatus);
          const onSlant = order.fulfillmentProvider === 'slant';
          const busy = updatingId === order.id;
          return (
            <div key={order.id} className="card space-y-3 p-5 text-sm">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-semibold text-reef-ink">Order {order.orderToken}</p>
                  <p className="text-reef-ink/60">{order.customerEmail}</p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="pill bg-reef-teal/10 text-reef-teal">
                    {onSlant ? `Slant: ${order.fulfillmentStatus || 'submitted'}` : order.fulfillmentStatus || 'submitted'}
                  </span>
                  {onSlant ? (
                    <button
                      onClick={() => handleSlantAction(order, 'refresh')}
                      disabled={busy}
                      className="btn-primary px-3 py-1.5 text-xs"
                    >
                      {busy ? 'Checking…' : 'Refresh Slant status'}
                    </button>
                  ) : (
                    <>
                      {action && (
                        <button
                          onClick={() => handleAdvance(order, action.next)}
                          disabled={busy}
                          className="btn-primary px-3 py-1.5 text-xs"
                        >
                          {busy ? 'Saving…' : action.label}
                        </button>
                      )}
                      {order.status !== 'fulfilled' && (
                        <button
                          onClick={() => handleSlantAction(order, 'send')}
                          disabled={busy}
                          className="rounded-full border-2 border-reef-ink px-3 py-1.5 text-xs font-bold text-reef-ink transition-colors hover:bg-reef-ink hover:text-white disabled:opacity-50"
                        >
                          {busy ? 'Sending…' : 'Send to Slant'}
                        </button>
                      )}
                    </>
                  )}
                </div>
              </div>

              <p className="text-reef-ink/70">Ship to: {shipToLine(order)}</p>

              <ul className="divide-y divide-reef-teal/10 border-t border-reef-teal/10">
                {order.items.map((item) => (
                  <li key={item.id} className="flex items-center justify-between py-2">
                    <span>
                      {item.productName || 'Item'}
                      {item.variantKey ? ` (${item.variantKey})` : ''} × {item.quantity}
                    </span>
                    <span className="flex items-center gap-3">
                      <span className="text-reef-ink/70">{usd(item.unitPriceCents * item.quantity)}</span>
                      {item.stlUrl && (
                        <a
                          href={item.stlUrl}
                          download
                          className="font-medium text-reef-teal underline underline-offset-2"
                        >
                          Download STL
                        </a>
                      )}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </div>
    </div>
  );
}
