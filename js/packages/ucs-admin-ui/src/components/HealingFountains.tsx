import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';

type HealingFountainRecord = {
  id: string;
  name: string;
  description: string;
  thumbnailUrl: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  invalidated?: boolean;
  zone?: {
    id: string;
    name: string;
  };
};

type GenerateHealingFountainImageResponse = {
  status?: string;
  thumbnailUrl?: string;
  healingFountain?: HealingFountainRecord;
  prompt?: string;
};

export const HealingFountains = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const [records, setRecords] = useState<HealingFountainRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [generatingId, setGeneratingId] = useState<string | null>(null);
  const [promptsById, setPromptsById] = useState<Record<string, string>>({});

  const zoneNameById = useMemo(() => {
    const map = new Map<string, string>();
    zones.forEach((zone) => map.set(zone.id, zone.name || zone.id));
    return map;
  }, [zones]);

  const fetchHealingFountains = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<HealingFountainRecord[]>(
        '/sonar/healing-fountains'
      );
      setRecords(response);
    } catch (err) {
      console.error('Failed to load healing fountains', err);
      setError('Failed to load healing fountains.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    fetchHealingFountains();
  }, [fetchHealingFountains]);

  const handleGenerateDiscoveredImage = async (record: HealingFountainRecord) => {
    if (generatingId) return;
    setGeneratingId(record.id);
    setError(null);
    try {
      const prompt = (promptsById[record.id] || '').trim();
      const response = await apiClient.post<GenerateHealingFountainImageResponse>(
        `/sonar/healing-fountains/${record.id}/generate-image`,
        prompt ? { prompt } : {}
      );
      const updatedRecord = response?.healingFountain;
      if (updatedRecord) {
        setRecords((prev) =>
          prev.map((entry) => (entry.id === updatedRecord.id ? updatedRecord : entry))
        );
      } else {
        await fetchHealingFountains();
      }
    } catch (err) {
      console.error('Failed to generate healing fountain image', err);
      setError('Failed to generate healing fountain image.');
    } finally {
      setGeneratingId(null);
    }
  };

  if (loading) {
    return <div className="m-10">Loading healing fountains...</div>;
  }

  return (
    <div className="m-10 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Healing Fountains</h1>
        <button
          type="button"
          onClick={() => fetchHealingFountains()}
          className="rounded bg-blue-600 px-3 py-2 text-white hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      <p className="text-sm text-gray-600">
        Generate the discovered thumbnail for each healing fountain. Undiscovered
        fountains use the same mystery icon/UX as points of interest in gameplay.
      </p>

      {error && (
        <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      {records.length === 0 ? (
        <div className="rounded border border-gray-200 bg-white p-6 text-sm text-gray-600">
          No healing fountains found.
        </div>
      ) : (
        <div className="grid gap-4">
          {records.map((record) => {
            const zoneName =
              record.zone?.name || zoneNameById.get(record.zoneId) || record.zoneId;
            const isGenerating = generatingId === record.id;
            return (
              <article
                key={record.id}
                className="grid gap-4 rounded border border-gray-200 bg-white p-4 shadow-sm md:grid-cols-[160px_1fr]"
              >
                <div className="h-40 w-40 overflow-hidden rounded border border-gray-200 bg-gray-50">
                  {record.thumbnailUrl?.trim() ? (
                    <img
                      src={record.thumbnailUrl}
                      alt={`${record.name || 'Healing Fountain'} thumbnail`}
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-xs text-gray-400">
                      No thumbnail
                    </div>
                  )}
                </div>
                <div className="space-y-3">
                  <div>
                    <h2 className="text-lg font-semibold">
                      {record.name || 'Healing Fountain'}
                    </h2>
                    <p className="text-sm text-gray-600">
                      Zone: {zoneName}
                    </p>
                    <p className="text-xs text-gray-500">
                      {record.latitude.toFixed(6)}, {record.longitude.toFixed(6)}
                    </p>
                  </div>
                  <p className="text-sm text-gray-700">
                    {record.description?.trim() || 'No description'}
                  </p>
                  <div className="space-y-2">
                    <label className="block text-xs font-medium text-gray-700">
                      Custom generation prompt (optional)
                    </label>
                    <textarea
                      value={promptsById[record.id] || ''}
                      onChange={(event) =>
                        setPromptsById((prev) => ({
                          ...prev,
                          [record.id]: event.target.value,
                        }))
                      }
                      rows={3}
                      className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      placeholder="Leave blank to use the default discovered healing fountain prompt."
                    />
                  </div>
                  <div>
                    <button
                      type="button"
                      onClick={() => handleGenerateDiscoveredImage(record)}
                      disabled={isGenerating}
                      className="rounded bg-emerald-600 px-3 py-2 text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-emerald-300"
                    >
                      {isGenerating ? 'Generating...' : 'Generate Discovered Image'}
                    </button>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default HealingFountains;
