import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

type ExpositionSeedProfileResponse = {
  category: string;
  label: string;
  firstSpawnChanceBasisPoints: number;
  secondSpawnChanceBasisPoints: number;
};

type ExpositionSeedConfigResponse = {
  id: number;
  profiles: ExpositionSeedProfileResponse[];
};

type EditableProfile = {
  category: string;
  label: string;
  firstSpawnChancePercent: number;
  secondSpawnChancePercent: number;
};

const basisPointsToPercent = (basisPoints: number) => {
  if (!Number.isFinite(basisPoints)) return 0;
  return Number((basisPoints / 100).toFixed(2));
};

const percentToBasisPoints = (percent: number) => {
  if (!Number.isFinite(percent)) return 0;
  const clamped = Math.max(0, Math.min(100, percent));
  return Math.round(clamped * 100);
};

const extractApiErrorMessage = (error: unknown, fallback: string): string => {
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    typeof (error as { response?: unknown }).response === 'object'
  ) {
    const response = (error as { response?: { data?: unknown } }).response;
    const data = response?.data;
    if (typeof data === 'object' && data !== null) {
      const maybeMessage = (data as { error?: unknown; message?: unknown })
        .error;
      if (typeof maybeMessage === 'string' && maybeMessage.trim() !== '') {
        return maybeMessage;
      }
      const maybeFallback = (data as { message?: unknown }).message;
      if (typeof maybeFallback === 'string' && maybeFallback.trim() !== '') {
        return maybeFallback;
      }
    }
  }
  if (error instanceof Error && error.message.trim() !== '') {
    return error.message;
  }
  return fallback;
};

const mapResponseToProfiles = (
  profiles: ExpositionSeedProfileResponse[]
): EditableProfile[] => {
  return (profiles || []).map((profile) => ({
    category: profile.category,
    label: profile.label,
    firstSpawnChancePercent: basisPointsToPercent(
      profile.firstSpawnChanceBasisPoints ?? 0
    ),
    secondSpawnChancePercent: basisPointsToPercent(
      profile.secondSpawnChanceBasisPoints ?? 0
    ),
  }));
};

export const PoiExpositionSeedConfigPanel = () => {
  const { apiClient } = useAPI();
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [profiles, setProfiles] = useState<EditableProfile[]>([]);

  const loadConfig = useCallback(
    async (showMessage = false) => {
      try {
        setLoading(true);
        setError(null);
        const response = await apiClient.get<ExpositionSeedConfigResponse>(
          '/sonar/admin/point-of-interest-exposition-seed-config'
        );
        setProfiles(mapResponseToProfiles(response?.profiles || []));
        if (showMessage) {
          setMessage('Reloaded saved exposition ratios.');
        }
      } catch (nextError) {
        console.error(
          'Failed to load point of interest exposition seed config',
          nextError
        );
        setError(
          extractApiErrorMessage(
            nextError,
            'Failed to load point of interest exposition seed config.'
          )
        );
      } finally {
        setLoading(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void loadConfig();
  }, [loadConfig]);

  const handleSave = useCallback(async () => {
    try {
      setSaving(true);
      setError(null);
      setMessage(null);

      const payload = {
        profiles: profiles.map((profile) => ({
          category: profile.category,
          firstSpawnChanceBasisPoints: percentToBasisPoints(
            profile.firstSpawnChancePercent
          ),
          secondSpawnChanceBasisPoints: percentToBasisPoints(
            profile.secondSpawnChancePercent
          ),
        })),
      };

      const response = await apiClient.put<ExpositionSeedConfigResponse>(
        '/sonar/admin/point-of-interest-exposition-seed-config',
        payload
      );
      setProfiles(mapResponseToProfiles(response?.profiles || []));
      setMessage('Point of interest exposition ratios saved.');
    } catch (nextError) {
      console.error(
        'Failed to save point of interest exposition seed config',
        nextError
      );
      setError(
        extractApiErrorMessage(
          nextError,
          'Failed to save point of interest exposition ratios.'
        )
      );
    } finally {
      setSaving(false);
    }
  }, [apiClient, profiles]);

  return (
    <div className="mb-6 rounded-lg bg-white p-4 shadow-md">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div className="max-w-3xl">
          <h2 className="text-lg font-semibold text-gray-900">
            POI Exposition Ratios
          </h2>
          <p className="mt-1 text-sm text-gray-600">
            Configure how likely each point of interest category is to spawn a
            first and second exposition during zone seeding.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setOpen((previous) => !previous)}
            className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            {open ? 'Hide Editor' : 'Show Editor'}
          </button>
          {open && (
            <>
              <button
                type="button"
                onClick={() => void loadConfig(true)}
                disabled={loading || saving}
                className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Reload Saved
              </button>
              <button
                type="button"
                onClick={() => void handleSave()}
                disabled={loading || saving}
                className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {saving ? 'Saving...' : 'Save Ratios'}
              </button>
            </>
          )}
        </div>
      </div>

      {error && (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}
      {message && (
        <div className="mt-4 rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
          {message}
        </div>
      )}

      {open && (
        <div className="mt-4">
          {loading ? (
            <div className="text-sm text-gray-500">
              Loading exposition ratios...
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
              {profiles.map((profile) => (
                <div
                  key={profile.category}
                  className="rounded-lg border border-gray-200 bg-gray-50 p-4"
                >
                  <div>
                    <h3 className="text-base font-semibold text-gray-900">
                      {profile.label}
                    </h3>
                    <p className="text-xs uppercase tracking-wide text-gray-500">
                      {profile.category}
                    </p>
                  </div>

                  <div className="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">
                        First Exposition (%)
                      </label>
                      <input
                        type="number"
                        min="0"
                        max="100"
                        step="0.1"
                        value={profile.firstSpawnChancePercent}
                        onChange={(event) =>
                          setProfiles((previous) =>
                            previous.map((current) =>
                              current.category === profile.category
                                ? {
                                    ...current,
                                    firstSpawnChancePercent:
                                      Number.parseFloat(event.target.value) || 0,
                                  }
                                : current
                            )
                          )
                        }
                        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">
                        Second Exposition (%)
                      </label>
                      <input
                        type="number"
                        min="0"
                        max="100"
                        step="0.1"
                        value={profile.secondSpawnChancePercent}
                        onChange={(event) =>
                          setProfiles((previous) =>
                            previous.map((current) =>
                              current.category === profile.category
                                ? {
                                    ...current,
                                    secondSpawnChancePercent:
                                      Number.parseFloat(event.target.value) || 0,
                                  }
                                : current
                            )
                          )
                        }
                        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PoiExpositionSeedConfigPanel;
