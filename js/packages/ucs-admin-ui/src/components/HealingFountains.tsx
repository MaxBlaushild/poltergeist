import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import ContentDashboard from './ContentDashboard.tsx';
import { countBy } from './contentDashboardUtils.ts';
import {
  useZoneKinds,
  zoneKindDescription,
  zoneKindLabel,
  zoneKindSelectPlaceholderLabel,
  zoneKindSummaryLabel,
} from './zoneKindHelpers.ts';
import { ContentMapMarkersMovedNotice } from './ContentMapMarkersMovedNotice.tsx';

type HealingFountainRecord = {
  id: string;
  name: string;
  description: string;
  thumbnailUrl: string;
  zoneId: string;
  zoneKind?: string;
  latitude: number;
  longitude: number;
  invalidated?: boolean;
  zone?: {
    id: string;
    name: string;
  };
};

type HealingFountainFormState = {
  name: string;
  description: string;
  thumbnailUrl: string;
  zoneId: string;
  zoneKind: string;
  latitude: string;
  longitude: string;
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

const defaultHealingFountainDiscoveredIconPrompt =
  'A discovered magical healing fountain in a retro 16-bit RPG style. Top-down map-ready icon art, luminous water, ancient stone basin, mystic runes, no text, no logos, centered composition, crisp outlines, limited palette.';

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

const emptyHealingFountainForm = (): HealingFountainFormState => ({
  name: '',
  description: '',
  thumbnailUrl: '',
  zoneId: '',
  zoneKind: '',
  latitude: '',
  longitude: '',
});

export const HealingFountains = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { zoneKindBySlug, zoneKinds } = useZoneKinds();
  const [records, setRecords] = useState<HealingFountainRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingRecord, setEditingRecord] =
    useState<HealingFountainRecord | null>(null);
  const [formData, setFormData] = useState<HealingFountainFormState>(
    emptyHealingFountainForm()
  );
  const [formError, setFormError] = useState<string | null>(null);
  const [savingRecord, setSavingRecord] = useState(false);

  const [discoveredIconPrompt, setDiscoveredIconPrompt] = useState(
    defaultHealingFountainDiscoveredIconPrompt
  );
  const [discoveredIconUrl, setDiscoveredIconUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/healing-fountain-discovered.png'
  );
  const [discoveredIconStatus, setDiscoveredIconStatus] =
    useState<string>('unknown');
  const [discoveredIconExists, setDiscoveredIconExists] = useState(false);
  const [discoveredIconRequestedAt, setDiscoveredIconRequestedAt] = useState<
    string | null
  >(null);
  const [discoveredIconLastModified, setDiscoveredIconLastModified] = useState<
    string | null
  >(null);
  const [discoveredIconStatusLoading, setDiscoveredIconStatusLoading] =
    useState(false);
  const [discoveredIconBusy, setDiscoveredIconBusy] = useState(false);
  const [discoveredIconMessage, setDiscoveredIconMessage] = useState<
    string | null
  >(null);
  const [discoveredIconError, setDiscoveredIconError] = useState<string | null>(
    null
  );
  const [discoveredIconPreviewNonce, setDiscoveredIconPreviewNonce] = useState(
    Date.now()
  );

  const zoneNameById = useMemo(() => {
    const map = new Map<string, string>();
    zones.forEach((zone) => map.set(zone.id, zone.name || zone.id));
    return map;
  }, [zones]);
  const zoneDefaultKindById = useMemo(() => {
    const map = new Map<string, string>();
    zones.forEach((zone) => map.set(zone.id, zone.kind?.trim() ?? ''));
    return map;
  }, [zones]);
  const selectedEditZoneDefaultKind = useMemo(
    () => zoneDefaultKindById.get(formData.zoneId) ?? '',
    [formData.zoneId, zoneDefaultKindById]
  );
  const editZoneKindDescription = useMemo(
    () =>
      zoneKindDescription(
        formData.zoneKind,
        selectedEditZoneDefaultKind,
        zoneKindBySlug
      ),
    [formData.zoneKind, selectedEditZoneDefaultKind, zoneKindBySlug]
  );
  const dashboardMetrics = useMemo(() => {
    const totalFountains = records.length;
    const invalidatedCount = records.filter((record) => record.invalidated)
      .length;
    const activeCount = totalFountains - invalidatedCount;
    const uniqueZones = new Set(records.map((record) => record.zoneId)).size;

    return [
      { label: 'Fountains', value: totalFountains },
      { label: 'Active', value: activeCount },
      { label: 'Invalidated', value: invalidatedCount },
      { label: 'Zones Covered', value: uniqueZones },
    ];
  }, [records]);
  const dashboardSections = useMemo(
    () => [
      {
        title: 'Zone Kinds',
        note: 'Effective healing fountain placement by zone kind.',
        buckets: countBy(records, (record) =>
          zoneKindLabel(
            record.zoneKind?.trim() || zoneDefaultKindById.get(record.zoneId),
            zoneKindBySlug
          )
        ),
      },
      {
        title: 'State',
        note: 'Whether fountains are currently active or invalidated.',
        buckets: countBy(records, (record) =>
          record.invalidated ? 'Invalidated' : 'Active'
        ),
      },
      {
        title: 'Zones',
        note: 'Top zones with healing fountain coverage.',
        buckets: countBy(records, (record) =>
          record.zone?.name || zoneNameById.get(record.zoneId) || record.zoneId
        ),
      },
    ],
    [records, zoneDefaultKindById, zoneKindBySlug, zoneNameById]
  );

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

  const refreshDiscoveredIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setDiscoveredIconStatusLoading(true);
        setDiscoveredIconError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/healing-fountain-discovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setDiscoveredIconUrl(url);
        }
        setDiscoveredIconStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setDiscoveredIconExists(Boolean(response?.exists));
        setDiscoveredIconRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setDiscoveredIconLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setDiscoveredIconPreviewNonce(Date.now());
        if (showMessage) {
          setDiscoveredIconMessage(
            'Discovered healing fountain icon status refreshed.'
          );
        }
      } catch (err) {
        console.error(
          'Failed to load discovered healing fountain icon status',
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load discovered healing fountain icon status.';
        setDiscoveredIconError(message);
      } finally {
        setDiscoveredIconStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateDiscoveredIcon = useCallback(async () => {
    const prompt = discoveredIconPrompt.trim();
    if (!prompt) {
      setDiscoveredIconError('Prompt is required.');
      return;
    }
    try {
      setDiscoveredIconBusy(true);
      setDiscoveredIconError(null);
      setDiscoveredIconMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/healing-fountain-discovered',
        { prompt }
      );
      setDiscoveredIconMessage(
        'Discovered healing fountain icon queued for generation.'
      );
      await refreshDiscoveredIconStatus();
    } catch (err) {
      console.error('Failed to generate discovered healing fountain icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate discovered healing fountain icon.';
      setDiscoveredIconError(message);
    } finally {
      setDiscoveredIconBusy(false);
    }
  }, [apiClient, discoveredIconPrompt, refreshDiscoveredIconStatus]);

  const handleDeleteDiscoveredIcon = useCallback(async () => {
    try {
      setDiscoveredIconBusy(true);
      setDiscoveredIconError(null);
      setDiscoveredIconMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/healing-fountain-discovered'
      );
      setDiscoveredIconMessage('Discovered healing fountain icon deleted.');
      await refreshDiscoveredIconStatus();
    } catch (err) {
      console.error('Failed to delete discovered healing fountain icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete discovered healing fountain icon.';
      setDiscoveredIconError(message);
    } finally {
      setDiscoveredIconBusy(false);
    }
  }, [apiClient, refreshDiscoveredIconStatus]);

  const handleEditRecord = useCallback((record: HealingFountainRecord) => {
    setEditingRecord(record);
    setFormError(null);
    setFormData({
      name: record.name ?? '',
      description: record.description ?? '',
      thumbnailUrl: record.thumbnailUrl ?? '',
      zoneId: record.zoneId ?? '',
      zoneKind: record.zoneKind ?? '',
      latitude: String(record.latitude ?? ''),
      longitude: String(record.longitude ?? ''),
    });
  }, []);

  const handleCloseEditModal = useCallback(() => {
    setEditingRecord(null);
    setFormError(null);
    setFormData(emptyHealingFountainForm());
  }, []);

  const handleSaveRecord = useCallback(async () => {
    if (!editingRecord) return;

    const zoneId = formData.zoneId.trim();
    const thumbnailUrl = formData.thumbnailUrl.trim();
    const latitude = Number.parseFloat(formData.latitude);
    const longitude = Number.parseFloat(formData.longitude);

    if (!zoneId) {
      setFormError('Zone is required.');
      return;
    }
    if (!thumbnailUrl) {
      setFormError('Thumbnail URL is required.');
      return;
    }
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      setFormError('Latitude and longitude are required.');
      return;
    }

    try {
      setSavingRecord(true);
      setFormError(null);
      const updated = await apiClient.put<HealingFountainRecord>(
        `/sonar/healing-fountains/${editingRecord.id}`,
        {
          name: formData.name.trim(),
          description: formData.description.trim(),
          thumbnailUrl,
          zoneId,
          zoneKind: formData.zoneKind,
          latitude,
          longitude,
        }
      );
      setRecords((prev) =>
        prev.map((record) =>
          record.id === updated.id
            ? {
                ...record,
                ...updated,
                zone: updated.zone ?? record.zone,
              }
            : record
        )
      );
      handleCloseEditModal();
    } catch (err) {
      console.error('Failed to update healing fountain', err);
      setFormError(
        err instanceof Error
          ? err.message
          : 'Failed to update healing fountain.'
      );
    } finally {
      setSavingRecord(false);
    }
  }, [apiClient, editingRecord, formData, handleCloseEditModal]);

  useEffect(() => {
    void fetchHealingFountains();
  }, [fetchHealingFountains]);

  useEffect(() => {
    void refreshDiscoveredIconStatus();
  }, [refreshDiscoveredIconStatus]);

  useEffect(() => {
    if (
      discoveredIconStatus !== 'queued' &&
      discoveredIconStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshDiscoveredIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [discoveredIconStatus, refreshDiscoveredIconStatus]);

  if (loading) {
    return <div className="m-10">Loading healing fountains...</div>;
  }

  return (
    <div className="m-10 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Healing Fountains</h1>
        <button
          type="button"
          onClick={() => void fetchHealingFountains()}
          className="rounded bg-blue-600 px-3 py-2 text-white hover:bg-blue-700"
        >
          Refresh Fountains
        </button>
      </div>

      <p className="text-sm text-gray-600">
        Discovered healing fountains use one shared S3 icon. Undiscovered
        healing fountains use the same mystery icon and UX as points of
        interest.
      </p>

      <ContentDashboard
        title="Healing Fountain Dashboard"
        subtitle="Aggregate healing fountain coverage across the world map."
        status="All healing fountains"
        metrics={dashboardMetrics}
        sections={dashboardSections}
      />

      <ContentMapMarkersMovedNotice subject="Healing fountain markers" />

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
              record.zone?.name ||
              zoneNameById.get(record.zoneId) ||
              record.zoneId;
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
                <div className="space-y-2">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h2 className="text-lg font-semibold">
                        {record.name || 'Healing Fountain'}
                      </h2>
                      <p className="text-sm text-gray-600">Zone: {zoneName}</p>
                      <p className="text-sm text-gray-600">
                        Zone Kind:{' '}
                        {zoneKindSummaryLabel(
                          record.zoneKind,
                          zoneDefaultKindById.get(record.zoneId) ?? '',
                          zoneKindBySlug
                        )}
                      </p>
                      {record.zoneKind?.trim() &&
                      (zoneDefaultKindById.get(record.zoneId) ?? '') &&
                      record.zoneKind.trim() !==
                        (zoneDefaultKindById.get(record.zoneId) ?? '') ? (
                        <p className="text-xs text-gray-500">
                          Zone default:{' '}
                          {zoneKindLabel(
                            zoneDefaultKindById.get(record.zoneId) ?? '',
                            zoneKindBySlug
                          )}
                        </p>
                      ) : null}
                      <p className="text-xs text-gray-500">
                        {record.latitude.toFixed(6)},{' '}
                        {record.longitude.toFixed(6)}
                      </p>
                    </div>
                    <button
                      type="button"
                      className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                      onClick={() => handleEditRecord(record)}
                    >
                      Edit
                    </button>
                  </div>
                  <p className="text-sm text-gray-700">
                    {record.description?.trim() || 'No description'}
                  </p>
                  <p className="text-xs text-gray-500 break-all">
                    Resolved thumbnail URL: {record.thumbnailUrl || 'n/a'}
                  </p>
                </div>
              </article>
            );
          })}
        </div>
      )}

      {editingRecord ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-xl font-bold">Edit Healing Fountain</h2>
              <button
                type="button"
                className="rounded border border-gray-300 bg-white px-3 py-2 text-sm hover:bg-gray-50"
                onClick={handleCloseEditModal}
              >
                Cancel
              </button>
            </div>

            {formError ? (
              <div className="mt-4 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {formError}
              </div>
            ) : null}

            <div className="mt-4 grid gap-4 md:grid-cols-2">
              <label className="block text-sm">
                Name
                <input
                  type="text"
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.name}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      name: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block text-sm">
                Zone
                <select
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.zoneId}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      zoneId: event.target.value,
                    }))
                  }
                >
                  <option value="">Select a zone</option>
                  {zones.map((zone) => (
                    <option
                      key={`healing-fountain-zone-${zone.id}`}
                      value={zone.id}
                    >
                      {zone.name || zone.id}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm">
                Zone Kind
                <select
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.zoneKind}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      zoneKind: event.target.value,
                    }))
                  }
                >
                  <option value="">
                    {zoneKindSelectPlaceholderLabel(
                      selectedEditZoneDefaultKind,
                      zoneKindBySlug
                    )}
                  </option>
                  {zoneKinds.map((zoneKind) => (
                    <option
                      key={`healing-fountain-zone-kind-${zoneKind.id}`}
                      value={zoneKind.slug}
                    >
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
                {editZoneKindDescription ? (
                  <p className="mt-1 text-xs text-gray-500">
                    {editZoneKindDescription}
                  </p>
                ) : null}
              </label>

              <label className="block text-sm">
                Thumbnail URL
                <input
                  type="text"
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.thumbnailUrl}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      thumbnailUrl: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block text-sm">
                Latitude
                <input
                  type="number"
                  step="any"
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.latitude}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      latitude: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block text-sm">
                Longitude
                <input
                  type="number"
                  step="any"
                  className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
                  value={formData.longitude}
                  onChange={(event) =>
                    setFormData((prev) => ({
                      ...prev,
                      longitude: event.target.value,
                    }))
                  }
                />
              </label>
            </div>

            <label className="mt-4 block text-sm">
              Description
              <textarea
                className="mt-1 min-h-[120px] w-full rounded border border-gray-300 px-3 py-2"
                value={formData.description}
                onChange={(event) =>
                  setFormData((prev) => ({
                    ...prev,
                    description: event.target.value,
                  }))
                }
              />
            </label>

            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                className="rounded border border-gray-300 bg-white px-4 py-2 text-sm font-medium hover:bg-gray-50"
                onClick={handleCloseEditModal}
              >
                Cancel
              </button>
              <button
                type="button"
                className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                onClick={() => void handleSaveRecord()}
                disabled={savingRecord}
              >
                {savingRecord ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default HealingFountains;
