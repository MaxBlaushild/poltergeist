import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

type BaseRecord = {
  id: string;
  userId: string;
  latitude: number;
  longitude: number;
  description?: string;
  imageUrl?: string;
  thumbnailUrl?: string;
  createdAt?: string;
  updatedAt?: string;
  owner?: {
    id?: string;
    name?: string;
    username?: string;
    profilePictureUrl?: string;
  };
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
};

type BaseDescriptionGenerationJob = {
  id: string;
  baseId: string;
  status?: string;
  generatedDescription?: string;
  errorMessage?: string;
  createdAt?: string;
  updatedAt?: string;
};

const defaultBaseIconPrompt =
  'A discovered adventurer base marker in a retro 16-bit fantasy MMORPG style. Top-down map-ready icon art, sturdy camp or homestead sigil, welcoming hearth glow, no text, no logos, centered composition, crisp outlines, limited palette.';

const staticStatusClassName = (status?: string) => {
  switch ((status || '').toLowerCase()) {
    case 'completed':
      return 'bg-emerald-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'queued':
      return 'bg-blue-600';
    case 'failed':
      return 'bg-red-600';
    case 'missing':
      return 'bg-slate-500';
    default:
      return 'bg-slate-500';
  }
};

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const ownerLabel = (record: BaseRecord) => {
  const username = record.owner?.username?.trim();
  if (username) return `@${username}`;
  const name = record.owner?.name?.trim();
  if (name) return name;
  return record.userId;
};

const secondaryOwnerLabel = (record: BaseRecord) => {
  const username = record.owner?.username?.trim();
  const name = record.owner?.name?.trim();
  if (!username || !name || name === username) return '';
  return name;
};

export const Bases = () => {
  const { apiClient } = useAPI();
  const [records, setRecords] = useState<BaseRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [iconPrompt, setIconPrompt] = useState(defaultBaseIconPrompt);
  const [iconUrl, setIconUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/base-discovered.png'
  );
  const [iconStatus, setIconStatus] = useState<string>('unknown');
  const [iconExists, setIconExists] = useState(false);
  const [iconRequestedAt, setIconRequestedAt] = useState<string | null>(null);
  const [iconLastModified, setIconLastModified] = useState<string | null>(null);
  const [iconStatusLoading, setIconStatusLoading] = useState(false);
  const [iconBusy, setIconBusy] = useState(false);
  const [iconMessage, setIconMessage] = useState<string | null>(null);
  const [iconError, setIconError] = useState<string | null>(null);
  const [iconPreviewNonce, setIconPreviewNonce] = useState(Date.now());
  const [isIconLightboxOpen, setIsIconLightboxOpen] = useState(false);
  const [deletingBaseId, setDeletingBaseId] = useState<string | null>(null);
  const [regeneratingBaseId, setRegeneratingBaseId] = useState<string | null>(null);
  const [baseMessage, setBaseMessage] = useState<string | null>(null);
  const [descriptionJobsByBaseId, setDescriptionJobsByBaseId] = useState<
    Record<string, BaseDescriptionGenerationJob>
  >({});

  const fetchBases = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<BaseRecord[]>('/sonar/admin/bases');
      setRecords(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load bases', err);
      setError('Failed to load bases.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  const refreshIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setIconStatusLoading(true);
        setIconError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/base/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setIconUrl(url);
        }
        setIconStatus((response?.status || 'unknown').trim() || 'unknown');
        setIconExists(Boolean(response?.exists));
        setIconRequestedAt(response?.requestedAt ? response.requestedAt : null);
        setIconLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setIconPreviewNonce(Date.now());
        if (showMessage) {
          setIconMessage('Base icon status refreshed.');
        }
      } catch (err) {
        console.error('Failed to load base icon status', err);
        const message =
          err instanceof Error ? err.message : 'Failed to load base icon status.';
        setIconError(message);
      } finally {
        setIconStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateIcon = useCallback(async () => {
    const prompt = iconPrompt.trim();
    if (!prompt) {
      setIconError('Prompt is required.');
      return;
    }
    try {
      setIconBusy(true);
      setIconError(null);
      setIconMessage(null);
      await apiClient.post('/sonar/admin/thumbnails/base', { prompt });
      setIconMessage('Base icon queued for generation.');
      await refreshIconStatus();
    } catch (err) {
      console.error('Failed to generate base icon', err);
      const message =
        err instanceof Error ? err.message : 'Failed to generate base icon.';
      setIconError(message);
    } finally {
      setIconBusy(false);
    }
  }, [apiClient, iconPrompt, refreshIconStatus]);

  const handleDeleteIcon = useCallback(async () => {
    try {
      setIconBusy(true);
      setIconError(null);
      setIconMessage(null);
      await apiClient.delete('/sonar/admin/thumbnails/base');
      setIconMessage('Base icon deleted.');
      await refreshIconStatus();
    } catch (err) {
      console.error('Failed to delete base icon', err);
      const message =
        err instanceof Error ? err.message : 'Failed to delete base icon.';
      setIconError(message);
    } finally {
      setIconBusy(false);
    }
  }, [apiClient, refreshIconStatus]);

  const handleDeleteBase = useCallback(
    async (record: BaseRecord) => {
      const owner = ownerLabel(record);
      const confirmed = window.confirm(
        `Delete ${owner}'s base? This will remove the base pin from the map.`
      );
      if (!confirmed) return;

      try {
        setDeletingBaseId(record.id);
        setError(null);
        setBaseMessage(null);
        await apiClient.delete(`/sonar/admin/bases/${record.id}`);
        setRecords((prev) => prev.filter((base) => base.id !== record.id));
        setBaseMessage(`Deleted ${owner}'s base.`);
      } catch (err) {
        console.error('Failed to delete base', err);
        const message =
          err instanceof Error ? err.message : 'Failed to delete base.';
        setError(message);
      } finally {
        setDeletingBaseId(null);
      }
    },
    [apiClient]
  );

  const fetchDescriptionJobs = useCallback(async () => {
    try {
      const response = await apiClient.get<BaseDescriptionGenerationJob[]>(
        '/sonar/admin/base-description-jobs?limit=100'
      );
      const jobs = Array.isArray(response) ? response : [];
      const next: Record<string, BaseDescriptionGenerationJob> = {};
      jobs.forEach((job) => {
        if (!job.baseId) return;
        if (!next[job.baseId]) {
          next[job.baseId] = job;
        }
      });
      setDescriptionJobsByBaseId(next);
    } catch (err) {
      console.error('Failed to load base description jobs', err);
    }
  }, [apiClient]);

  const handleRegenerateDescription = useCallback(
    async (record: BaseRecord) => {
      try {
        setRegeneratingBaseId(record.id);
        setError(null);
        setBaseMessage(null);
        await apiClient.post(`/sonar/admin/bases/${record.id}/generate-description`);
        setBaseMessage(`Queued base flavor generation for ${ownerLabel(record)}.`);
        await fetchDescriptionJobs();
      } catch (err) {
        console.error('Failed to queue base description generation', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to queue base description generation.';
        setError(message);
      } finally {
        setRegeneratingBaseId(null);
      }
    },
    [apiClient, fetchDescriptionJobs]
  );

  useEffect(() => {
    void fetchBases();
  }, [fetchBases]);

  useEffect(() => {
    void refreshIconStatus();
  }, [refreshIconStatus]);

  useEffect(() => {
    void fetchDescriptionJobs();
  }, [fetchDescriptionJobs]);

  useEffect(() => {
    if (iconStatus !== 'queued' && iconStatus !== 'in_progress') {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [iconStatus, refreshIconStatus]);

  useEffect(() => {
    const hasPendingJobs = Object.values(descriptionJobsByBaseId).some((job) =>
      ['queued', 'in_progress'].includes((job.status || '').toLowerCase())
    );
    if (!hasPendingJobs) {
      return;
    }
    const interval = window.setInterval(() => {
      void fetchDescriptionJobs();
      void fetchBases();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [descriptionJobsByBaseId, fetchBases, fetchDescriptionJobs]);

  if (loading) {
    return <div className="m-10">Loading bases...</div>;
  }

  return (
    <div className="m-10 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Bases</h1>
        <button
          type="button"
          onClick={() => void fetchBases()}
          className="rounded bg-blue-600 px-3 py-2 text-white hover:bg-blue-700"
        >
          Refresh Bases
        </button>
      </div>

      <p className="text-sm text-gray-600">
        Bases are player-owned map pins created in the app. They all share one
        generated icon.
      </p>

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">Base Pin Icon</h2>
            <p className="mt-1 text-xs text-gray-600">
              Requested: {formatDate(iconRequestedAt ?? undefined)}
            </p>
            <p className="text-xs text-gray-600">
              Last updated: {formatDate(iconLastModified ?? undefined)}
            </p>
          </div>
          <span
            className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
              iconStatus
            )}`}
          >
            {iconStatus || 'unknown'}
          </span>
        </div>

        <div className="mt-4 flex flex-wrap gap-3">
          <button
            type="button"
            onClick={() => void refreshIconStatus(true)}
            disabled={iconStatusLoading}
            className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconStatusLoading ? 'Refreshing...' : 'Refresh Status'}
          </button>
          <button
            type="button"
            onClick={() => void handleGenerateIcon()}
            disabled={iconBusy || iconStatusLoading}
            className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconBusy ? 'Working...' : 'Generate Icon'}
          </button>
          <button
            type="button"
            onClick={() => void handleDeleteIcon()}
            disabled={iconBusy || iconStatusLoading}
            className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconBusy ? 'Working...' : 'Delete Icon'}
          </button>
        </div>

        <label className="mt-4 block text-sm font-medium text-gray-700">
          Prompt
          <textarea
            value={iconPrompt}
            onChange={(e) => setIconPrompt(e.target.value)}
            rows={4}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
          />
        </label>

        <p className="mt-3 text-xs text-gray-600 break-all">URL: {iconUrl}</p>

        {iconExists ? (
          <div className="mt-4 flex justify-center rounded border border-dashed border-gray-300 bg-gray-50 p-4">
            <button
              type="button"
              onClick={() => setIsIconLightboxOpen(true)}
              className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              title="Open large preview"
            >
              <img
                src={`${iconUrl}?v=${iconPreviewNonce}`}
                alt="Base icon preview"
                className="h-28 w-28 rounded object-contain"
              />
            </button>
          </div>
        ) : (
          <div className="mt-4 rounded border border-dashed border-gray-300 bg-gray-50 p-4 text-sm text-gray-500">
            No generated icon found yet.
          </div>
        )}

        {iconMessage ? (
          <p className="mt-3 text-sm text-emerald-700">{iconMessage}</p>
        ) : null}
        {iconError ? <p className="mt-3 text-sm text-red-700">{iconError}</p> : null}
      </section>

      {isIconLightboxOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-6"
          onClick={() => setIsIconLightboxOpen(false)}
        >
          <div
            className="relative max-h-[90vh] max-w-[90vw] rounded-lg bg-white p-4 shadow-2xl"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setIsIconLightboxOpen(false)}
              className="absolute right-3 top-3 rounded bg-black/70 px-2 py-1 text-xs font-semibold text-white hover:bg-black/80"
            >
              Close
            </button>
            <img
              src={`${iconUrl}?v=${iconPreviewNonce}`}
              alt="Large base icon preview"
              className="max-h-[80vh] max-w-[80vw] rounded object-contain"
            />
          </div>
        </div>
      ) : null}

      {error ? (
        <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}
      {baseMessage ? (
        <div className="rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {baseMessage}
        </div>
      ) : null}

      <section className="space-y-3">
        {records.length === 0 ? (
          <div className="rounded border border-gray-200 bg-white px-4 py-6 text-sm text-gray-500 shadow-sm">
            No bases created yet.
          </div>
        ) : (
          records.map((record) => {
            const latestJob = descriptionJobsByBaseId[record.id];
            return (
              <div
                key={record.id}
                className="rounded border border-gray-200 bg-white p-4 shadow-sm"
              >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 className="text-sm font-semibold text-gray-900">
                    {ownerLabel(record)}
                  </h3>
                  {secondaryOwnerLabel(record) ? (
                    <p className="text-xs text-gray-500">
                      {secondaryOwnerLabel(record)}
                    </p>
                  ) : null}
                </div>
                <div className="text-xs text-gray-500">
                  Updated {formatDate(record.updatedAt)}
                </div>
              </div>
              {record.thumbnailUrl?.trim() ? (
                <div className="mt-3">
                  <img
                    src={record.thumbnailUrl}
                    alt={`${ownerLabel(record)} base`}
                    className="h-28 w-28 rounded object-cover"
                  />
                </div>
              ) : null}
              {record.description?.trim() ? (
                <p className="mt-3 text-sm leading-6 text-gray-700">
                  {record.description.trim()}
                </p>
              ) : (
                <p className="mt-3 text-sm italic text-gray-500">
                  No description generated yet.
                </p>
              )}
              {latestJob ? (
                <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
                  <span
                    className={`rounded-full px-2 py-1 font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                      latestJob.status
                    )}`}
                  >
                    {latestJob.status || 'unknown'}
                  </span>
                  <span className="text-gray-500">
                    {formatDate(latestJob.updatedAt)}
                  </span>
                  {latestJob.errorMessage ? (
                    <span className="text-red-700">{latestJob.errorMessage}</span>
                  ) : null}
                </div>
              ) : null}
              <div className="mt-3 grid gap-2 text-sm text-gray-700 md:grid-cols-3">
                <div>Latitude: {record.latitude.toFixed(6)}</div>
                <div>Longitude: {record.longitude.toFixed(6)}</div>
                <div>User ID: {record.userId}</div>
              </div>
              <div className="mt-4 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => void handleRegenerateDescription(record)}
                  disabled={regeneratingBaseId === record.id}
                  className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {regeneratingBaseId === record.id
                    ? 'Queueing...'
                    : 'Regenerate Flavor'}
                </button>
                <button
                  type="button"
                  onClick={() => void handleDeleteBase(record)}
                  disabled={deletingBaseId === record.id}
                  className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {deletingBaseId === record.id ? 'Deleting...' : 'Delete Base'}
                </button>
              </div>
              </div>
            );
          })
        )}
      </section>
    </div>
  );
};

export default Bases;
