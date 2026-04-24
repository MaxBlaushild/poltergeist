import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { useZoneKinds } from './zoneKindHelpers.ts';

type SharedMarkerStatus = {
  id: string;
  label: string;
  thumbnailUrl: string;
  effectiveThumbnailUrl: string;
  defaultThumbnailUrl: string;
  status: string;
  exists: boolean;
  defaultExists: boolean;
  requestedAt?: string | null;
  lastModified?: string | null;
  defaultPrompt: string;
  actionPath: string;
  supportsZoneKinds: boolean;
  zoneKind?: string;
};

type PoiCategoryMarkerStatus = {
  category: string;
  label: string;
  defaultPrompt: string;
  thumbnailUrl: string;
  effectiveThumbnailUrl: string;
  defaultThumbnailUrl: string;
  status: string;
  exists: boolean;
  defaultExists: boolean;
  actionPath: string;
  supportsZoneKinds: boolean;
  zoneKind?: string;
  requestedAt?: string | null;
  lastModified?: string | null;
};

type ResourceTypeMarkerStatus = {
  resourceTypeId: string;
  name: string;
  slug: string;
  description: string;
  thumbnailUrl: string;
  effectiveThumbnailUrl: string;
  defaultThumbnailUrl: string;
  status: string;
  exists: boolean;
  defaultExists: boolean;
  requestedAt?: string | null;
  lastModified?: string | null;
  defaultPrompt: string;
  zoneKind?: string;
  supportsZoneKinds: boolean;
  canDelete: boolean;
};

type ContentMapMarkersResponse = {
  sharedMarkers: SharedMarkerStatus[];
  poiCategoryMarkers: PoiCategoryMarkerStatus[];
  resourceTypeMarkers: ResourceTypeMarkerStatus[];
  zoneKind?: {
    slug: string;
    name: string;
    description: string;
  } | null;
};

const statusClassName = (status?: string) => {
  const normalized = (status || '').trim().toLowerCase();
  if (normalized === 'completed') return 'bg-emerald-600';
  if (normalized === 'queued' || normalized === 'in_progress')
    return 'bg-indigo-600';
  if (normalized === 'failed' || normalized === 'missing') return 'bg-red-600';
  return 'bg-slate-500';
};

const formatDate = (value?: string | null) => {
  if (!value) return '—';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const buildQueryString = (zoneKind: string) => {
  const params = new URLSearchParams();
  if (zoneKind.trim()) {
    params.set('zoneKind', zoneKind.trim());
  }
  const query = params.toString();
  return query ? `?${query}` : '';
};

const resolvePreviewUrl = (entry: {
  exists: boolean;
  thumbnailUrl: string;
  defaultExists: boolean;
  defaultThumbnailUrl: string;
  effectiveThumbnailUrl: string;
}) => {
  if (entry.exists && entry.thumbnailUrl.trim()) return entry.thumbnailUrl;
  if (entry.defaultExists && entry.defaultThumbnailUrl.trim()) {
    return entry.defaultThumbnailUrl;
  }
  return entry.effectiveThumbnailUrl.trim();
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
      const value = (data as { error?: unknown; message?: unknown }).error;
      if (typeof value === 'string' && value.trim()) return value;
      const message = (data as { message?: unknown }).message;
      if (typeof message === 'string' && message.trim()) return message;
    }
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
};

const markerCardClassName =
  'rounded-lg border border-slate-200 bg-white p-4 shadow-sm';

const sectionTitleClassName = 'text-lg font-semibold text-slate-900';

const ContentMapMarkers = () => {
  const { apiClient } = useAPI();
  const { zoneKinds } = useZoneKinds();
  const [selectedZoneKind, setSelectedZoneKind] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [busyKey, setBusyKey] = useState<string | null>(null);
  const [data, setData] = useState<ContentMapMarkersResponse | null>(null);
  const [sharedPrompts, setSharedPrompts] = useState<Record<string, string>>(
    {}
  );
  const [poiPrompts, setPoiPrompts] = useState<Record<string, string>>({});
  const [resourcePrompts, setResourcePrompts] = useState<Record<string, string>>(
    {}
  );

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<ContentMapMarkersResponse>(
        `/sonar/admin/content-map-markers${buildQueryString(selectedZoneKind)}`
      );
      setData(response);
      setSharedPrompts(
        Object.fromEntries(
          (response.sharedMarkers || []).map((entry) => [
            entry.id,
            entry.defaultPrompt || '',
          ])
        )
      );
      setPoiPrompts(
        Object.fromEntries(
          (response.poiCategoryMarkers || []).map((entry) => [
            entry.category,
            entry.defaultPrompt || '',
          ])
        )
      );
      setResourcePrompts(
        Object.fromEntries(
          (response.resourceTypeMarkers || []).map((entry) => [
            entry.resourceTypeId,
            entry.defaultPrompt || '',
          ])
        )
      );
    } catch (nextError) {
      console.error('Failed to load content map markers page data', nextError);
      setError(
        extractApiErrorMessage(
          nextError,
          'Unable to load content map markers right now.'
        )
      );
      setData(null);
    } finally {
      setLoading(false);
    }
  }, [apiClient, selectedZoneKind]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const zoneScopedSharedMarkers = useMemo(
    () =>
      (data?.sharedMarkers || []).filter((entry) => entry.supportsZoneKinds),
    [data]
  );
  const defaultOnlySharedMarkers = useMemo(
    () =>
      (data?.sharedMarkers || []).filter((entry) => !entry.supportsZoneKinds),
    [data]
  );
  const activeZoneKindLabel = data?.zoneKind?.name?.trim() || 'Default';
  const zoneKindDescription = data?.zoneKind?.description?.trim() || '';
  const isDefaultView = !selectedZoneKind.trim();

  const handleSharedMarkerAction = useCallback(
    async (entry: SharedMarkerStatus, action: 'generate' | 'delete') => {
      const prompt = sharedPrompts[entry.id]?.trim() || entry.defaultPrompt;
      setBusyKey(`shared:${entry.id}:${action}`);
      setError(null);
      setMessage(null);
      try {
        const path = `${entry.actionPath}${buildQueryString(selectedZoneKind)}`;
        if (action === 'generate') {
          await apiClient.post(path, { prompt });
          setMessage(`Queued ${entry.label.toLowerCase()} marker generation.`);
        } else {
          await apiClient.delete(path);
          setMessage(`Removed the ${entry.label.toLowerCase()} override.`);
        }
        await loadData();
      } catch (nextError) {
        console.error(`Failed to ${action} shared marker`, nextError);
        setError(
          extractApiErrorMessage(
            nextError,
            `Unable to ${action} the ${entry.label.toLowerCase()} marker.`
          )
        );
      } finally {
        setBusyKey(null);
      }
    },
    [apiClient, loadData, selectedZoneKind, sharedPrompts]
  );

  const handlePoiCategoryAction = useCallback(
    async (
      entry: PoiCategoryMarkerStatus,
      action: 'generate' | 'delete'
    ) => {
      const prompt = poiPrompts[entry.category]?.trim() || entry.defaultPrompt;
      setBusyKey(`poi:${entry.category}:${action}`);
      setError(null);
      setMessage(null);
      try {
        const path = `${entry.actionPath}${buildQueryString(selectedZoneKind)}`;
        if (action === 'generate') {
          await apiClient.post(path, { prompt });
          setMessage(`Queued ${entry.label.toLowerCase()} marker generation.`);
        } else {
          await apiClient.delete(path);
          setMessage(`Removed the ${entry.label.toLowerCase()} override.`);
        }
        await loadData();
      } catch (nextError) {
        console.error(`Failed to ${action} POI category marker`, nextError);
        setError(
          extractApiErrorMessage(
            nextError,
            `Unable to ${action} the ${entry.label.toLowerCase()} marker.`
          )
        );
      } finally {
        setBusyKey(null);
      }
    },
    [apiClient, loadData, poiPrompts, selectedZoneKind]
  );

  const handleResourceMarkerAction = useCallback(
    async (
      entry: ResourceTypeMarkerStatus,
      action: 'generate' | 'delete'
    ) => {
      setBusyKey(`resource:${entry.resourceTypeId}:${action}`);
      setError(null);
      setMessage(null);
      try {
        if (action === 'generate') {
          const prompt =
            resourcePrompts[entry.resourceTypeId]?.trim() || entry.defaultPrompt;
          await apiClient.post(
            `/sonar/resource-types/${entry.resourceTypeId}/generate-map-icon${buildQueryString(
              selectedZoneKind
            )}`,
            { prompt }
          );
          setMessage(`Generated ${entry.name} marker art.`);
        } else {
          await apiClient.delete(
            `/sonar/admin/content-map-markers/resource-types/${entry.resourceTypeId}${buildQueryString(
              selectedZoneKind
            )}`
          );
          setMessage(`Removed the ${entry.name} zone-kind override.`);
        }
        await loadData();
      } catch (nextError) {
        console.error(`Failed to ${action} resource marker`, nextError);
        setError(
          extractApiErrorMessage(
            nextError,
            `Unable to ${action} the ${entry.name} marker.`
          )
        );
      } finally {
        setBusyKey(null);
      }
    },
    [apiClient, loadData, resourcePrompts, selectedZoneKind]
  );

  if (loading && !data) {
    return <div className="p-6 text-slate-500">Loading content markers...</div>;
  }

  return (
    <div className="space-y-6 p-6">
      <section className="rounded-2xl border border-slate-200 bg-gradient-to-br from-white via-slate-50 to-amber-50 p-6 shadow-sm">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-slate-950">
              Content Map Markers
            </h1>
            <p className="mt-2 max-w-3xl text-sm text-slate-600">
              Manage the default marker set and optional zone-kind overrides in
              one place. When an override exists, content in zones of that kind
              uses it. When it does not, the live map falls back to the default
              marker that already exists today.
            </p>
          </div>
          <button
            type="button"
            className="rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-700"
            onClick={() => void loadData()}
            disabled={loading}
          >
            {loading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>

        <div className="mt-5 grid gap-4 lg:grid-cols-[280px_minmax(0,1fr)]">
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-700">
              Marker Set
            </span>
            <select
              className="w-full rounded-md border border-slate-300 px-3 py-2"
              value={selectedZoneKind}
              onChange={(event) => {
                setSelectedZoneKind(event.target.value);
                setMessage(null);
              }}
            >
              <option value="">Default markers</option>
              {zoneKinds.map((zoneKind) => (
                <option key={zoneKind.id} value={zoneKind.slug}>
                  {zoneKind.name}
                </option>
              ))}
            </select>
          </label>

          <div className="rounded-xl border border-slate-200 bg-white/90 p-4">
            <div className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
              Active Target
            </div>
            <div className="mt-1 text-lg font-semibold text-slate-900">
              {activeZoneKindLabel}
            </div>
            <p className="mt-2 text-sm text-slate-600">
              {isDefaultView
                ? 'These are the current shared markers. Every zone falls back to this set unless it has a zone-kind-specific override.'
                : `Overrides for ${activeZoneKindLabel} zones. Missing entries fall back to the default markers automatically.`}
            </p>
            <p className="mt-2 text-sm text-slate-500">
              Override generation automatically combines the base content prompt
              with the selected zone kind&apos;s flavor, so a marker can stay
              legible while shifting into that biome&apos;s most iconic form.
            </p>
            {zoneKindDescription ? (
              <p className="mt-2 text-sm text-slate-500">{zoneKindDescription}</p>
            ) : null}
          </div>
        </div>
      </section>

      {error ? (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}
      {message ? (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {message}
        </div>
      ) : null}

      <section className="space-y-4">
        <div>
          <h2 className={sectionTitleClassName}>Shared Content Markers</h2>
          <p className="mt-1 text-sm text-slate-600">
            Scenarios, monster encounters, mystery POIs, healing fountains, and
            other shared pins.
          </p>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          {zoneScopedSharedMarkers.map((entry) => {
            const previewUrl = resolvePreviewUrl(entry);
            const usingDefaultFallback = !entry.exists && entry.defaultExists;
            const busy = busyKey?.startsWith(`shared:${entry.id}:`);
            return (
              <article key={entry.id} className={markerCardClassName}>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 className="text-base font-semibold text-slate-900">
                      {entry.label}
                    </h3>
                    <p className="mt-1 break-all text-xs text-slate-500">
                      Target URL: {entry.thumbnailUrl}
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      Requested: {formatDate(entry.requestedAt)}
                      {' · '}
                      Last updated: {formatDate(entry.lastModified)}
                    </p>
                  </div>
                  <span
                    className={`inline-flex rounded-full px-2 py-0.5 text-xs text-white ${statusClassName(
                      entry.status
                    )}`}
                  >
                    {entry.status || 'unknown'}
                  </span>
                </div>

                <label className="mt-3 block text-sm">
                  <span className="mb-1 block font-medium text-slate-700">
                    Generation Prompt
                  </span>
                  <textarea
                    className="min-h-[96px] w-full rounded-md border border-slate-300 p-2"
                    value={sharedPrompts[entry.id] ?? entry.defaultPrompt}
                    onChange={(event) =>
                      setSharedPrompts((current) => ({
                        ...current,
                        [entry.id]: event.target.value,
                      }))
                    }
                  />
                </label>

                <div className="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-60"
                    onClick={() => void handleSharedMarkerAction(entry, 'generate')}
                    disabled={Boolean(busy)}
                  >
                    {busy ? 'Working…' : isDefaultView ? 'Generate Default' : 'Generate Override'}
                  </button>
                  <button
                    type="button"
                    className="rounded-md border border-red-200 px-3 py-2 text-sm text-red-700 disabled:opacity-60"
                    onClick={() => void handleSharedMarkerAction(entry, 'delete')}
                    disabled={Boolean(busy) || isDefaultView}
                  >
                    Delete Override
                  </button>
                </div>

                {previewUrl ? (
                  <div className="mt-4 flex items-center gap-4">
                    <img
                      src={previewUrl}
                      alt={`${entry.label} preview`}
                      className="h-20 w-20 rounded-lg border border-slate-200 bg-slate-50 object-cover"
                    />
                    <div className="text-xs text-slate-500">
                      {entry.exists
                        ? 'Using the active marker override.'
                        : usingDefaultFallback
                        ? 'No override found yet. The live map will use the default marker.'
                        : 'No generated marker found yet for this slot.'}
                    </div>
                  </div>
                ) : (
                  <p className="mt-3 text-xs text-slate-500">
                    No preview is available yet.
                  </p>
                )}
              </article>
            );
          })}
        </div>

        {defaultOnlySharedMarkers.length > 0 ? (
          <div className="rounded-xl border border-dashed border-slate-300 bg-white p-4">
            <div className="text-sm font-semibold text-slate-900">
              Default-Only Markers
            </div>
            <p className="mt-1 text-sm text-slate-600">
              These markers are global and do not currently vary by zone kind.
            </p>
            <div className="mt-4 grid gap-4 lg:grid-cols-2">
              {defaultOnlySharedMarkers.map((entry) => {
                const previewUrl = resolvePreviewUrl(entry);
                const busy = busyKey?.startsWith(`shared:${entry.id}:`);
                return (
                  <article key={entry.id} className={markerCardClassName}>
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <h3 className="text-base font-semibold text-slate-900">
                          {entry.label}
                        </h3>
                        <p className="mt-1 break-all text-xs text-slate-500">
                          URL: {entry.thumbnailUrl}
                        </p>
                      </div>
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs text-white ${statusClassName(
                          entry.status
                        )}`}
                      >
                        {entry.status || 'unknown'}
                      </span>
                    </div>
                    <label className="mt-3 block text-sm">
                      <span className="mb-1 block font-medium text-slate-700">
                        Generation Prompt
                      </span>
                      <textarea
                        className="min-h-[96px] w-full rounded-md border border-slate-300 p-2"
                        value={sharedPrompts[entry.id] ?? entry.defaultPrompt}
                        onChange={(event) =>
                          setSharedPrompts((current) => ({
                            ...current,
                            [entry.id]: event.target.value,
                          }))
                        }
                      />
                    </label>
                    <div className="mt-3 flex flex-wrap gap-2">
                      <button
                        type="button"
                        className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-60"
                        onClick={() =>
                          void handleSharedMarkerAction(entry, 'generate')
                        }
                        disabled={Boolean(busy)}
                      >
                        {busy ? 'Working…' : 'Generate Default'}
                      </button>
                    </div>
                    {previewUrl ? (
                      <div className="mt-4">
                        <img
                          src={previewUrl}
                          alt={`${entry.label} preview`}
                          className="h-20 w-20 rounded-lg border border-slate-200 bg-slate-50 object-cover"
                        />
                      </div>
                    ) : null}
                  </article>
                );
              })}
            </div>
          </div>
        ) : null}
      </section>

      <section className="space-y-4">
        <div>
          <h2 className={sectionTitleClassName}>POI Category Markers</h2>
          <p className="mt-1 text-sm text-slate-600">
            Discovered point-of-interest icons by marker category.
          </p>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          {(data?.poiCategoryMarkers || []).map((entry) => {
            const previewUrl = resolvePreviewUrl(entry);
            const busy = busyKey?.startsWith(`poi:${entry.category}:`);
            return (
              <article key={entry.category} className={markerCardClassName}>
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="text-base font-semibold text-slate-900">
                      {entry.label}
                    </h3>
                    <p className="mt-1 break-all text-xs text-slate-500">
                      Target URL: {entry.thumbnailUrl}
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      Requested: {formatDate(entry.requestedAt)}
                      {' · '}
                      Last updated: {formatDate(entry.lastModified)}
                    </p>
                  </div>
                  <span
                    className={`inline-flex rounded-full px-2 py-0.5 text-xs text-white ${statusClassName(
                      entry.status
                    )}`}
                  >
                    {entry.status || 'unknown'}
                  </span>
                </div>

                <label className="mt-3 block text-sm">
                  <span className="mb-1 block font-medium text-slate-700">
                    Generation Prompt
                  </span>
                  <textarea
                    className="min-h-[96px] w-full rounded-md border border-slate-300 p-2"
                    value={poiPrompts[entry.category] ?? entry.defaultPrompt}
                    onChange={(event) =>
                      setPoiPrompts((current) => ({
                        ...current,
                        [entry.category]: event.target.value,
                      }))
                    }
                  />
                </label>

                <div className="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-60"
                    onClick={() => void handlePoiCategoryAction(entry, 'generate')}
                    disabled={Boolean(busy)}
                  >
                    {busy ? 'Working…' : isDefaultView ? 'Generate Default' : 'Generate Override'}
                  </button>
                  <button
                    type="button"
                    className="rounded-md border border-red-200 px-3 py-2 text-sm text-red-700 disabled:opacity-60"
                    onClick={() => void handlePoiCategoryAction(entry, 'delete')}
                    disabled={Boolean(busy) || isDefaultView}
                  >
                    Delete Override
                  </button>
                </div>

                {previewUrl ? (
                  <div className="mt-4 flex items-center gap-4">
                    <img
                      src={previewUrl}
                      alt={`${entry.label} preview`}
                      className="h-20 w-20 rounded-lg border border-slate-200 bg-slate-50 object-cover"
                    />
                    <div className="text-xs text-slate-500">
                      {entry.exists
                        ? 'Using the active category marker override.'
                        : entry.defaultExists
                        ? 'No override found yet. The game will fall back to the default category marker.'
                        : 'If no generated asset exists, the client still falls back to the built-in marker art.'}
                    </div>
                  </div>
                ) : (
                  <p className="mt-3 text-xs text-slate-500">
                    No generated preview is available yet.
                  </p>
                )}
              </article>
            );
          })}
        </div>
      </section>

      <section className="space-y-4">
        <div>
          <h2 className={sectionTitleClassName}>Resource Type Markers</h2>
          <p className="mt-1 text-sm text-slate-600">
            Resource type icons use the selected zone-kind override when it
            exists, and otherwise fall back to the stored default marker for the
            resource type.
          </p>
        </div>
        <div className="space-y-4">
          {(data?.resourceTypeMarkers || []).map((entry) => {
            const busy = busyKey?.startsWith(`resource:${entry.resourceTypeId}:`);
            const previewUrl =
              entry.effectiveThumbnailUrl.trim() || entry.defaultThumbnailUrl;
            return (
              <article
                key={entry.resourceTypeId}
                className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
              >
                <div className="grid gap-4 lg:grid-cols-[96px_minmax(0,1fr)]">
                  <div className="flex items-start justify-center">
                    {previewUrl ? (
                      <img
                        src={previewUrl}
                        alt={`${entry.name} marker preview`}
                        className="h-24 w-24 rounded-lg border border-slate-200 bg-slate-50 object-contain"
                      />
                    ) : (
                      <div className="flex h-24 w-24 items-center justify-center rounded-lg border border-dashed border-slate-300 bg-slate-50 text-xs text-slate-400">
                        No icon
                      </div>
                    )}
                  </div>

                  <div>
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h3 className="text-base font-semibold text-slate-900">
                          {entry.name}
                        </h3>
                        <p className="text-xs uppercase tracking-[0.18em] text-slate-500">
                          {entry.slug}
                        </p>
                        <p className="mt-2 text-sm text-slate-600">
                          {entry.description || 'No description yet.'}
                        </p>
                        <p className="mt-2 break-all text-xs text-slate-500">
                          Effective URL:{' '}
                          {entry.effectiveThumbnailUrl || entry.defaultThumbnailUrl || '—'}
                        </p>
                      </div>
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs text-white ${statusClassName(
                          entry.status
                        )}`}
                      >
                        {entry.status || 'unknown'}
                      </span>
                    </div>

                    <label className="mt-3 block text-sm">
                      <span className="mb-1 block font-medium text-slate-700">
                        Generation Prompt
                      </span>
                      <textarea
                        className="min-h-[96px] w-full rounded-md border border-slate-300 p-2"
                        value={
                          resourcePrompts[entry.resourceTypeId] ??
                          entry.defaultPrompt
                        }
                        onChange={(event) =>
                          setResourcePrompts((current) => ({
                            ...current,
                            [entry.resourceTypeId]: event.target.value,
                          }))
                        }
                      />
                    </label>

                    <div className="mt-3 flex flex-wrap gap-2">
                      <button
                        type="button"
                        className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-60"
                        onClick={() =>
                          void handleResourceMarkerAction(entry, 'generate')
                        }
                        disabled={Boolean(busy)}
                      >
                        {busy
                          ? 'Working…'
                          : isDefaultView
                          ? 'Generate Default'
                          : 'Generate Override'}
                      </button>
                      <button
                        type="button"
                        className="rounded-md border border-red-200 px-3 py-2 text-sm text-red-700 disabled:opacity-60"
                        onClick={() =>
                          void handleResourceMarkerAction(entry, 'delete')
                        }
                        disabled={Boolean(busy) || !entry.canDelete}
                      >
                        Delete Override
                      </button>
                    </div>

                    {!entry.exists && entry.defaultExists && !isDefaultView ? (
                      <p className="mt-3 text-xs text-slate-500">
                        No zone-kind override exists yet. Live zones will keep
                        using the default resource marker until one is generated.
                      </p>
                    ) : null}
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      </section>
    </div>
  );
};

export default ContentMapMarkers;
