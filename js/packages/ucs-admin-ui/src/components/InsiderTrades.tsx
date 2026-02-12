import { useAPI } from '@poltergeist/contexts';
import React, { useEffect, useState } from 'react';

interface InsiderTrade {
  id: string;
  externalId: string;
  marketId?: string;
  marketName?: string;
  outcome?: string;
  side?: string;
  price?: number;
  size?: number;
  notional?: number;
  trader?: string;
  tradeTime?: string;
  detectedAt?: string;
  reason?: string;
}

export const InsiderTrades = () => {
  const { apiClient } = useAPI();
  const [trades, setTrades] = useState<InsiderTrade[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTrades = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<InsiderTrade[]>(
        '/sonar/admin/insider-trades'
      );
      setTrades(Array.isArray(response) ? response : []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(`Failed to load insider trades: ${msg}`);
      setTrades([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTrades();
  }, []);

  if (loading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Insider Trades</h1>
        <p>Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Insider Trades</h1>
        <p className="text-red-600">{error}</p>
        <button
          onClick={fetchTrades}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Insider Trades</h1>
        <button
          onClick={fetchTrades}
          className="px-3 py-2 text-sm bg-gray-800 text-white rounded hover:bg-gray-700"
        >
          Refresh
        </button>
      </div>
      <p className="text-gray-600 mb-4">
        Trades flagged by the basic anomaly detector. Use these for review only.
      </p>

      {trades.length === 0 ? (
        <p className="text-gray-500">No insider trades detected.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full border text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="p-2 border">Detected</th>
                <th className="p-2 border">Market</th>
                <th className="p-2 border">Side</th>
                <th className="p-2 border">Outcome</th>
                <th className="p-2 border">Size</th>
                <th className="p-2 border">Price</th>
                <th className="p-2 border">Notional</th>
                <th className="p-2 border">Trader</th>
                <th className="p-2 border">Reason</th>
              </tr>
            </thead>
            <tbody>
              {trades.map((trade) => (
                <tr key={trade.id} className="odd:bg-white even:bg-gray-50">
                  <td className="p-2 border whitespace-nowrap">
                    {trade.detectedAt ? new Date(trade.detectedAt).toLocaleString() : '--'}
                  </td>
                  <td className="p-2 border">
                    <div className="font-medium">
                      {trade.marketName || trade.marketId || 'Unknown'}
                    </div>
                    {trade.marketId && (
                      <div className="text-xs text-gray-500">{trade.marketId}</div>
                    )}
                  </td>
                  <td className="p-2 border">{trade.side || '--'}</td>
                  <td className="p-2 border">{trade.outcome || '--'}</td>
                  <td className="p-2 border text-right">
                    {trade.size !== undefined ? trade.size.toFixed(2) : '--'}
                  </td>
                  <td className="p-2 border text-right">
                    {trade.price !== undefined ? trade.price.toFixed(4) : '--'}
                  </td>
                  <td className="p-2 border text-right">
                    {trade.notional !== undefined ? trade.notional.toFixed(2) : '--'}
                  </td>
                  <td className="p-2 border">{trade.trader || '--'}</td>
                  <td className="p-2 border">{trade.reason || '--'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
