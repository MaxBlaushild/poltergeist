import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

type FeedbackItem = {
  id: string;
  createdAt: string;
  updatedAt: string;
  route: string;
  message: string;
  userId: string;
  zoneId?: string | null;
  user?: {
    id: string;
    name?: string;
    username?: string | null;
    phoneNumber?: string;
  } | null;
  zone?: {
    id: string;
    name?: string;
  } | null;
};

const formatDate = (value?: string) => {
  if (!value) return '--';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

export const Feedback = () => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<FeedbackItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFeedback = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<FeedbackItem[]>('/sonar/admin/feedback');
      setItems(Array.isArray(response) ? response : []);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setError(`Failed to load feedback: ${message}`);
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void fetchFeedback();
  }, [fetchFeedback]);

  return (
    <div className="p-6">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Feedback</h1>
          <p className="text-sm text-gray-600">
            In-app shake submissions from players.
          </p>
        </div>
        <button
          onClick={fetchFeedback}
          className="rounded bg-gray-800 px-3 py-2 text-sm text-white hover:bg-gray-700"
        >
          Refresh
        </button>
      </div>

      {loading ? (
        <p>Loading...</p>
      ) : error ? (
        <p className="text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-gray-500">No feedback submitted yet.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full border text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="border p-2 text-left">Created</th>
                <th className="border p-2 text-left">User</th>
                <th className="border p-2 text-left">Zone</th>
                <th className="border p-2 text-left">Route</th>
                <th className="border p-2 text-left">Message</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => {
                const userLabel =
                  item.user?.username?.trim() ||
                  item.user?.name?.trim() ||
                  item.userId;
                return (
                  <tr key={item.id} className="odd:bg-white even:bg-gray-50">
                    <td className="border p-2 whitespace-nowrap">
                      {formatDate(item.createdAt)}
                    </td>
                    <td className="border p-2">
                      <div className="font-medium">{userLabel}</div>
                      {item.user?.phoneNumber && (
                        <div className="text-xs text-gray-500">{item.user?.phoneNumber}</div>
                      )}
                    </td>
                    <td className="border p-2">{item.zone?.name || item.zoneId || '--'}</td>
                    <td className="border p-2 whitespace-nowrap">{item.route || '--'}</td>
                    <td className="border p-2 whitespace-pre-wrap">{item.message}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default Feedback;
