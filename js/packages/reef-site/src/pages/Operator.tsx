import { useEffect, useState } from 'react';
import { reefApi } from '../api/client';
import type { OperatorMetrics } from '../api/types';

function pct(ratio: number): string {
  return `${(ratio * 100).toFixed(1)}%`;
}

function usd(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

// R-9.2: the single operator view carrying the four go/no-go numbers —
// configurator-to-purchase conversion, validation rejection rate (by rule),
// mean landed COGS, and CAC (ad spend entered manually). Not linked from the
// storefront nav; reached directly at /operator.
export default function Operator() {
  const [days, setDays] = useState(30);
  const [adSpend, setAdSpend] = useState('0');
  const [metrics, setMetrics] = useState<OperatorMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = (targetDays: number, adSpendCents: number) => {
    reefApi
      .operatorMetrics(targetDays, adSpendCents)
      .then(setMetrics)
      .catch(() => setError('Failed to load metrics'))
      .finally(() => setLoading(false));
  };

  const load = () => {
    setLoading(true);
    setError(null);
    fetchMetrics(days, Math.round(Number(adSpend) * 100) || 0);
  };

  useEffect(() => {
    fetchMetrics(30, 0);
  }, []);

  return (
    <div className="max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">Operator metrics</h1>

      <div className="flex flex-wrap items-end gap-4 text-sm">
        <label className="flex flex-col gap-1">
          Window (days)
          <input
            type="number"
            min={1}
            value={days}
            onChange={(e) => setDays(Number(e.target.value) || 30)}
            className="w-24 rounded border border-reef-teal/30 px-2 py-1"
          />
        </label>
        <label className="flex flex-col gap-1">
          Ad spend ($, manual)
          <input
            type="number"
            min={0}
            step="0.01"
            value={adSpend}
            onChange={(e) => setAdSpend(e.target.value)}
            className="w-32 rounded border border-reef-teal/30 px-2 py-1"
          />
        </label>
        <button
          onClick={load}
          disabled={loading}
          className="rounded bg-reef-coral px-4 py-2 font-semibold text-reef-ink hover:opacity-90 disabled:opacity-50"
        >
          {loading ? 'Loading…' : 'Refresh'}
        </button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}

      {metrics && (
        <div className="space-y-6">
          <section>
            <h2 className="font-semibold mb-2">Configurator-to-purchase conversion</h2>
            <dl className="grid grid-cols-3 gap-2 text-sm">
              <Stat label="Configurator opened" value={metrics.configuratorOpened} />
              <Stat label="Purchases completed" value={metrics.purchaseCompleted} />
              <Stat label="Conversion rate" value={pct(metrics.conversionRate)} />
            </dl>
          </section>

          <section>
            <h2 className="font-semibold mb-2">Validation rejection rate</h2>
            <dl className="grid grid-cols-3 gap-2 text-sm mb-3">
              <Stat label="Validated total" value={metrics.validatedTotal} />
              <Stat label="Rejected" value={metrics.rejectedTotal} />
              <Stat label="Rejection rate" value={pct(metrics.rejectionRate)} />
            </dl>
            {Object.keys(metrics.rejectionsByRule).length > 0 ? (
              <table className="w-full text-sm border-collapse">
                <tbody>
                  {Object.entries(metrics.rejectionsByRule).map(([rule, count]) => (
                    <tr key={rule} className="border-t border-reef-teal/10">
                      <td className="py-1 pr-4">{rule}</td>
                      <td className="py-1 text-right">{count}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className="text-sm text-reef-ink/50">No rejections in this window.</p>
            )}
          </section>

          <section>
            <h2 className="font-semibold mb-2">True landed COGS</h2>
            <dl className="grid grid-cols-3 gap-2 text-sm">
              <Stat label="Paid orders" value={metrics.paidOrderCount} />
              <Stat label="COGS recorded for" value={metrics.cogsRecordedForOrders} />
              <Stat label="Mean COGS" value={usd(metrics.meanCogsCents)} />
            </dl>
          </section>

          <section>
            <h2 className="font-semibold mb-2">Reprints</h2>
            <dl className="grid grid-cols-3 gap-2 text-sm">
              <Stat label="Orders reprinted" value={metrics.ordersReprinted} />
              <Stat label="Reprint rate" value={pct(metrics.reprintRate)} />
              <div />
            </dl>
          </section>

          <section>
            <h2 className="font-semibold mb-2">CAC</h2>
            <dl className="grid grid-cols-3 gap-2 text-sm">
              <Stat label="Ad spend" value={usd(metrics.adSpendCents)} />
              <Stat label="Purchases" value={metrics.purchaseCompleted} />
              <Stat label="CAC" value={usd(metrics.cacCents)} />
            </dl>
          </section>
        </div>
      )}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded border border-reef-teal/20 px-3 py-2">
      <dt className="text-xs text-reef-ink/50">{label}</dt>
      <dd className="text-lg font-semibold">{value}</dd>
    </div>
  );
}
