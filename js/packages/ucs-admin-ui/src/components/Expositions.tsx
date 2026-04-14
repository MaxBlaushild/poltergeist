import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import { useSearchParams } from 'react-router-dom';
import {
  Character,
  DialogueMessage,
  Exposition,
  ExpositionTemplate,
  InventoryItem,
  PointOfInterest,
  Spell,
} from '@poltergeist/types';
import { DialogueMessageListEditor } from './DialogueMessageListEditor.tsx';
import {
  MaterialRewardsEditor,
  MaterialRewardForm,
  normalizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';

type ExpositionFormState = {
  zoneId: string;
  locationMode: 'poi' | 'coordinates';
  pointOfInterestId: string;
  latitude: string;
  longitude: string;
  title: string;
  description: string;
  dialogue: DialogueMessage[];
  imageUrl: string;
  thumbnailUrl: string;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  rewardGold: string;
  materialRewards: MaterialRewardForm[];
  itemRewards: Array<{ inventoryItemId: string; quantity: number }>;
  spellRewards: Array<{ spellId: string }>;
};

type StaticThumbnailResponse = {
  status?: string;
  exists?: boolean;
  thumbnailUrl?: string;
  requestedAt?: string | null;
  lastModified?: string | null;
};

const emptyExpositionForm = (): ExpositionFormState => ({
  zoneId: '',
  locationMode: 'coordinates',
  pointOfInterestId: '',
  latitude: '',
  longitude: '',
  title: '',
  description: '',
  dialogue: [],
  imageUrl: '',
  thumbnailUrl: '',
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  rewardGold: '0',
  materialRewards: [],
  itemRewards: [],
  spellRewards: [],
});

const buildFormFromExposition = (record: Exposition): ExpositionFormState => ({
  zoneId: record.zoneId ?? '',
  locationMode: record.pointOfInterestId ? 'poi' : 'coordinates',
  pointOfInterestId: record.pointOfInterestId ?? '',
  latitude: String(record.latitude ?? ''),
  longitude: String(record.longitude ?? ''),
  title: record.title ?? '',
  description: record.description ?? '',
  dialogue: record.dialogue ?? [],
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  rewardExperience: String(record.rewardExperience ?? 0),
  rewardGold: String(record.rewardGold ?? 0),
  materialRewards: (record.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount ?? 1,
  })),
  itemRewards: (record.itemRewards ?? []).map((reward) => ({
    inventoryItemId: reward.inventoryItemId
      ? String(reward.inventoryItemId)
      : '',
    quantity: reward.quantity ?? 1,
  })),
  spellRewards: (record.spellRewards ?? []).map((reward) => ({
    spellId: reward.spellId ?? '',
  })),
});

const parsePositiveInt = (value: string, fallback = 0) => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
};

const defaultExpositionMapIconPrompt =
  'A retro 16-bit RPG map marker icon for an exposition encounter. Dialogue scroll sigil with multiple speaking silhouettes, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

const defaultExpositionMapIconUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/exposition-undiscovered.png';

const formatDate = (value?: string | null): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const staticStatusClassName = (status: string): string => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    case 'missing':
      return 'bg-gray-500';
    default:
      return 'bg-gray-400';
  }
};

export const Expositions: React.FC = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();
  const [searchParams, setSearchParams] = useSearchParams();
  const [records, setRecords] = useState<Exposition[]>([]);
  const [selectedId, setSelectedId] = useState<string>('');
  const [form, setForm] = useState<ExpositionFormState>(emptyExpositionForm());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [savingTemplate, setSavingTemplate] = useState(false);
  const [locationBusy, setLocationBusy] = useState(false);
  const [error, setError] = useState('');
  const [statusMessage, setStatusMessage] = useState('');
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>(
    []
  );
  const [characters, setCharacters] = useState<Character[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [expositionIconPrompt, setExpositionIconPrompt] = useState(
    defaultExpositionMapIconPrompt
  );
  const [expositionIconUrl, setExpositionIconUrl] = useState(
    defaultExpositionMapIconUrl
  );
  const [expositionIconStatus, setExpositionIconStatus] = useState('unknown');
  const [expositionIconExists, setExpositionIconExists] = useState(false);
  const [expositionIconRequestedAt, setExpositionIconRequestedAt] = useState<
    string | null
  >(null);
  const [expositionIconLastModified, setExpositionIconLastModified] = useState<
    string | null
  >(null);
  const [expositionIconStatusLoading, setExpositionIconStatusLoading] =
    useState(false);
  const [expositionIconBusy, setExpositionIconBusy] = useState(false);
  const [expositionIconMessage, setExpositionIconMessage] = useState<
    string | null
  >(null);
  const [expositionIconError, setExpositionIconError] = useState<string | null>(
    null
  );
  const [expositionIconPreviewNonce, setExpositionIconPreviewNonce] = useState(
    Date.now()
  );
  const deepLinkedId = searchParams.get('id')?.trim() ?? '';

  const selectedRecord = useMemo(
    () => records.find((record) => record.id === selectedId) ?? null,
    [records, selectedId]
  );

  const characterOptions = useMemo(
    () =>
      characters
        .map((character) => ({
          value: character.id,
          label: character.name?.trim() || character.id,
        }))
        .sort((a, b) => a.label.localeCompare(b.label)),
    [characters]
  );

  const sortedRecords = useMemo(
    () =>
      [...records].sort((a, b) =>
        (a.title || '').localeCompare(b.title || '', undefined, {
          sensitivity: 'base',
        })
      ),
    [records]
  );

  const refreshExpositionIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setExpositionIconStatusLoading(true);
        setExpositionIconError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/exposition-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setExpositionIconUrl(url);
        }
        setExpositionIconStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setExpositionIconExists(Boolean(response?.exists));
        setExpositionIconRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setExpositionIconLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setExpositionIconPreviewNonce(Date.now());
        if (showMessage) {
          setExpositionIconMessage('Exposition map icon status refreshed.');
        }
      } catch (err) {
        console.error('Failed to load exposition map icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load exposition map icon status.';
        setExpositionIconError(message);
      } finally {
        setExpositionIconStatusLoading(false);
      }
    },
    [apiClient]
  );

  const loadRecords = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await apiClient.get<{ items: Exposition[] }>(
        '/sonar/admin/expositions'
      );
      const items = response.items ?? [];
      setRecords(items);
      if (deepLinkedId && items.some((item) => item.id === deepLinkedId)) {
        setSelectedId(deepLinkedId);
        setForm(
          buildFormFromExposition(
            items.find((item) => item.id === deepLinkedId)!
          )
        );
        return;
      }
      if (items.length === 0) {
        setSelectedId('');
        setForm((prev) =>
          prev.zoneId
            ? prev
            : {
                ...emptyExpositionForm(),
                zoneId: zones[0]?.id ?? '',
              }
        );
      } else if (!items.some((item) => item.id === selectedId)) {
        setSelectedId(items[0].id);
        setForm(buildFormFromExposition(items[0]));
      }
    } catch (err) {
      console.error('Failed to load expositions', err);
      setError('Failed to load expositions.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadRecords();
    void refreshExpositionIconStatus();
    void (async () => {
      try {
        const [spellResponse, characterResponse] = await Promise.all([
          apiClient.get<Spell[]>('/sonar/spells'),
          apiClient.get<Character[]>('/sonar/characters'),
        ]);
        setSpells(spellResponse ?? []);
        setCharacters(characterResponse ?? []);
      } catch (err) {
        console.error('Failed to load exposition dependencies', err);
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (
      expositionIconStatus !== 'queued' &&
      expositionIconStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshExpositionIconStatus();
    }, 5000);
    return () => window.clearInterval(interval);
  }, [expositionIconStatus, refreshExpositionIconStatus]);

  useEffect(() => {
    if (selectedRecord) {
      setForm(buildFormFromExposition(selectedRecord));
      return;
    }
    setForm((prev) =>
      prev.zoneId
        ? prev
        : {
            ...emptyExpositionForm(),
            zoneId: zones[0]?.id ?? '',
          }
    );
  }, [selectedRecord, zones]);

  useEffect(() => {
    if (!deepLinkedId) return;
    if (!records.some((record) => record.id === deepLinkedId)) return;
    setSelectedId(deepLinkedId);
  }, [deepLinkedId, records]);

  useEffect(() => {
    if (!selectedId) return;
    if (searchParams.get('id') === selectedId) return;
    const next = new URLSearchParams(searchParams);
    next.set('id', selectedId);
    setSearchParams(next, { replace: true });
  }, [searchParams, selectedId, setSearchParams]);

  useEffect(() => {
    const zoneId = form.zoneId.trim();
    if (!zoneId) {
      setPointsOfInterest([]);
      return;
    }
    void (async () => {
      try {
        const response = await apiClient.get<PointOfInterest[]>(
          `/sonar/zones/${zoneId}/pointsOfInterest`
        );
        setPointsOfInterest(response ?? []);
      } catch (err) {
        console.error('Failed to load exposition POIs', err);
        setPointsOfInterest([]);
      }
    })();
  }, [apiClient, form.zoneId]);

  const resetForNew = () => {
    setSelectedId('');
    setStatusMessage('');
    setError('');
    const next = new URLSearchParams(searchParams);
    next.delete('id');
    setSearchParams(next, { replace: true });
    setForm({
      ...emptyExpositionForm(),
      zoneId: form.zoneId || zones[0]?.id || '',
    });
  };

  const validateForm = () => {
    if (!form.zoneId.trim()) {
      return 'Zone is required.';
    }
    if (!form.title.trim()) {
      return 'Title is required.';
    }
    if (form.locationMode === 'coordinates') {
      if (!form.latitude.trim() || !form.longitude.trim()) {
        return 'Latitude and longitude are required for coordinate placement.';
      }
    } else if (!form.pointOfInterestId.trim()) {
      return 'Pick a point of interest or switch to coordinates.';
    }
    const hasDialogue = form.dialogue.some(
      (line) => line.text.trim().length > 0
    );
    if (!hasDialogue) {
      return 'Dialogue is required.';
    }
    const missingCharacter = form.dialogue.some(
      (line) =>
        (line.speaker ?? 'character') === 'character' &&
        line.text.trim().length > 0 &&
        !(line.characterId ?? '').trim()
    );
    if (missingCharacter) {
      return 'Every character dialogue line needs a speaker selected.';
    }
    return '';
  };

  const buildPayload = () => ({
    zoneId: form.zoneId,
    pointOfInterestId:
      form.locationMode === 'poi' ? form.pointOfInterestId || null : null,
    latitude:
      form.locationMode === 'coordinates'
        ? Number.parseFloat(form.latitude) || 0
        : 0,
    longitude:
      form.locationMode === 'coordinates'
        ? Number.parseFloat(form.longitude) || 0
        : 0,
    title: form.title.trim(),
    description: form.description.trim(),
    dialogue: form.dialogue,
    imageUrl: form.imageUrl.trim(),
    thumbnailUrl: form.thumbnailUrl.trim(),
    rewardMode: form.rewardMode,
    randomRewardSize: form.randomRewardSize,
    rewardExperience:
      form.rewardMode === 'explicit'
        ? Number.parseInt(form.rewardExperience, 10) || 0
        : 0,
    rewardGold:
      form.rewardMode === 'explicit'
        ? Number.parseInt(form.rewardGold, 10) || 0
        : 0,
    materialRewards:
      form.rewardMode === 'explicit'
        ? normalizeMaterialRewards(form.materialRewards)
        : [],
    itemRewards:
      form.rewardMode === 'explicit'
        ? form.itemRewards
            .filter((reward) => reward.inventoryItemId)
            .map((reward) => ({
              inventoryItemId: Number.parseInt(reward.inventoryItemId, 10) || 0,
              quantity: parsePositiveInt(String(reward.quantity), 1),
            }))
        : [],
    spellRewards:
      form.rewardMode === 'explicit'
        ? form.spellRewards
            .filter((reward) => reward.spellId)
            .map((reward) => ({ spellId: reward.spellId }))
        : [],
  });

  const handleSave = async () => {
    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }
    setSaving(true);
    setError('');
    setStatusMessage('');
    try {
      const payload = buildPayload();
      const saved = selectedId
        ? await apiClient.put<Exposition>(
            `/sonar/expositions/${selectedId}`,
            payload
          )
        : await apiClient.post<Exposition>('/sonar/expositions', payload);
      setRecords((prev) => {
        const next = prev.filter((record) => record.id !== saved.id);
        next.push(saved);
        return next;
      });
      setSelectedId(saved.id);
      setForm(buildFormFromExposition(saved));
      setStatusMessage(
        selectedId ? 'Exposition updated.' : 'Exposition created.'
      );
    } catch (err) {
      console.error('Failed to save exposition', err);
      setError('Failed to save exposition.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedId) return;
    if (!window.confirm('Delete this exposition?')) return;
    setSaving(true);
    setError('');
    setStatusMessage('');
    try {
      await apiClient.delete(`/sonar/expositions/${selectedId}`);
      const nextRecords = records.filter((record) => record.id !== selectedId);
      setRecords(nextRecords);
      if (nextRecords.length > 0) {
        setSelectedId(nextRecords[0].id);
        setForm(buildFormFromExposition(nextRecords[0]));
      } else {
        resetForNew();
      }
      setStatusMessage('Exposition deleted.');
    } catch (err) {
      console.error('Failed to delete exposition', err);
      setError('Failed to delete exposition.');
    } finally {
      setSaving(false);
    }
  };

  const handleGenerateImage = async () => {
    if (!selectedId) return;
    setSaving(true);
    setError('');
    setStatusMessage('');
    try {
      await apiClient.post(
        `/sonar/expositions/${selectedId}/generate-image`,
        {}
      );
      setStatusMessage('Image generation queued.');
    } catch (err) {
      console.error('Failed to queue exposition image generation', err);
      setError('Failed to queue image generation.');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveTemplate = async () => {
    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }
    setSavingTemplate(true);
    setError('');
    setStatusMessage('');
    try {
      const template = await apiClient.post<ExpositionTemplate>(
        '/sonar/exposition-templates',
        buildPayload()
      );
      setStatusMessage(
        `Saved "${template.title || 'exposition'}" as a reusable exposition template.`
      );
    } catch (err) {
      console.error('Failed to save exposition template', err);
      setError('Failed to save exposition template.');
    } finally {
      setSavingTemplate(false);
    }
  };

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      setError('Geolocation is not supported in this browser.');
      return;
    }
    setLocationBusy(true);
    setError('');
    setStatusMessage('');
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLocationBusy(false);
        setForm((prev) => ({
          ...prev,
          locationMode: 'coordinates',
          latitude: String(position.coords.latitude),
          longitude: String(position.coords.longitude),
        }));
        setStatusMessage('Filled coordinates from your current location.');
      },
      (geoError) => {
        setLocationBusy(false);
        setError(`Unable to get current location: ${geoError.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, []);

  const handleGenerateExpositionIcon = useCallback(async () => {
    const prompt = expositionIconPrompt.trim();
    if (!prompt) {
      setExpositionIconError('Prompt is required.');
      return;
    }
    try {
      setExpositionIconBusy(true);
      setExpositionIconError(null);
      setExpositionIconMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/exposition-undiscovered',
        { prompt }
      );
      setExpositionIconMessage('Exposition map icon queued for generation.');
      await refreshExpositionIconStatus();
    } catch (err) {
      console.error('Failed to generate exposition map icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate exposition map icon.';
      setExpositionIconError(message);
    } finally {
      setExpositionIconBusy(false);
    }
  }, [apiClient, expositionIconPrompt, refreshExpositionIconStatus]);

  const handleDeleteExpositionIcon = useCallback(async () => {
    try {
      setExpositionIconBusy(true);
      setExpositionIconError(null);
      setExpositionIconMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/exposition-undiscovered'
      );
      setExpositionIconMessage('Exposition map icon deleted.');
      await refreshExpositionIconStatus();
    } catch (err) {
      console.error('Failed to delete exposition map icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete exposition map icon.';
      setExpositionIconError(message);
    } finally {
      setExpositionIconBusy(false);
    }
  }, [apiClient, refreshExpositionIconStatus]);

  const selectedPointOfInterest = useMemo(
    () =>
      pointsOfInterest.find((point) => point.id === form.pointOfInterestId) ??
      null,
    [form.pointOfInterestId, pointsOfInterest]
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-slate-900">Expositions</h1>
          <p className="mt-1 text-sm text-slate-600">
            Author dialogue-driven encounters that can live on the map or sit
            behind a quest node.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-900"
            onClick={resetForNew}
          >
            New Exposition
          </button>
          <button
            type="button"
            className="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
            onClick={handleSave}
            disabled={saving}
          >
            {saving
              ? 'Saving...'
              : selectedId
                ? 'Save Changes'
                : 'Create Exposition'}
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <div>
            <h2 className="text-lg font-semibold text-slate-900">
              Exposition Map Icon
            </h2>
            <p className="mt-1 text-sm text-slate-600">
              This shared static icon is used for expositions both outside the
              interaction radius and when players are close enough to engage.
            </p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              className="rounded-md bg-slate-700 px-3 py-1 text-white disabled:opacity-60"
              onClick={() => void refreshExpositionIconStatus(true)}
              disabled={expositionIconStatusLoading}
            >
              {expositionIconStatusLoading ? 'Refreshing…' : 'Refresh Status'}
            </button>
            <button
              type="button"
              className="rounded-md bg-indigo-600 px-3 py-1 text-white disabled:opacity-60"
              onClick={handleGenerateExpositionIcon}
              disabled={expositionIconBusy || expositionIconStatusLoading}
            >
              {expositionIconBusy ? 'Working…' : 'Generate Icon'}
            </button>
            <button
              type="button"
              className="rounded-md bg-red-600 px-3 py-1 text-white disabled:opacity-60"
              onClick={handleDeleteExpositionIcon}
              disabled={expositionIconBusy || expositionIconStatusLoading}
            >
              {expositionIconBusy ? 'Working…' : 'Delete Icon'}
            </button>
          </div>
        </div>
        <div className="mb-2">
          <span
            className={`inline-flex rounded px-2 py-0.5 text-xs text-white ${staticStatusClassName(
              expositionIconStatus
            )}`}
          >
            {expositionIconStatus || 'unknown'}
          </span>
        </div>
        <div className="break-all text-xs text-slate-600">
          URL: {expositionIconUrl}
        </div>
        <div className="mt-1 text-xs text-slate-600">
          Requested: {formatDate(expositionIconRequestedAt)}
          {' · '}
          Last updated: {formatDate(expositionIconLastModified)}
        </div>
        <label className="mt-3 block text-sm">
          Generation Prompt
          <textarea
            className="mt-1 min-h-[88px] w-full rounded-md border border-slate-300 p-2"
            value={expositionIconPrompt}
            onChange={(event) => setExpositionIconPrompt(event.target.value)}
            placeholder="Prompt used to generate the exposition map icon."
          />
        </label>
        {expositionIconExists ? (
          <div className="mt-3">
            <img
              src={`${expositionIconUrl}?v=${expositionIconPreviewNonce}`}
              alt="Exposition map icon preview"
              className="h-24 w-24 rounded-md border bg-slate-50 object-cover"
            />
          </div>
        ) : (
          <div className="mt-2 text-xs text-slate-500">
            No icon currently found at this URL.
          </div>
        )}
        {expositionIconMessage ? (
          <div className="mt-2 text-sm text-emerald-700">
            {expositionIconMessage}
          </div>
        ) : null}
        {expositionIconError ? (
          <div className="mt-2 text-sm text-red-600">{expositionIconError}</div>
        ) : null}
      </div>

      {error ? (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}
      {statusMessage ? (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {statusMessage}
        </div>
      ) : null}

      <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <aside className="space-y-3 rounded-xl border border-slate-200 bg-white p-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">
              Existing Expositions
            </h2>
            <button
              type="button"
              className="text-sm text-slate-600 underline"
              onClick={() => void loadRecords()}
            >
              Refresh
            </button>
          </div>
          {loading ? (
            <div className="text-sm text-slate-500">Loading…</div>
          ) : sortedRecords.length === 0 ? (
            <div className="rounded-md border border-dashed border-slate-300 px-3 py-4 text-sm text-slate-500">
              No expositions yet.
            </div>
          ) : (
            <div className="space-y-2">
              {sortedRecords.map((record) => {
                const active = record.id === selectedId;
                const zoneName =
                  zones.find((zone) => zone.id === record.zoneId)?.name ??
                  record.zoneId;
                return (
                  <button
                    key={record.id}
                    type="button"
                    onClick={() => setSelectedId(record.id)}
                    className={`w-full rounded-lg border px-3 py-3 text-left transition ${
                      active
                        ? 'border-slate-900 bg-slate-900 text-white'
                        : 'border-slate-200 bg-slate-50 text-slate-900 hover:border-slate-400'
                    }`}
                  >
                    <div className="text-sm font-semibold">
                      {record.title || '(Untitled exposition)'}
                    </div>
                    <div
                      className={`mt-1 text-xs ${
                        active ? 'text-slate-200' : 'text-slate-500'
                      }`}
                    >
                      {zoneName}
                    </div>
                    <div
                      className={`mt-2 text-xs ${
                        active ? 'text-slate-200' : 'text-slate-600'
                      }`}
                    >
                      {record.dialogue?.length ?? 0} lines
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </aside>

        <section className="space-y-6 rounded-xl border border-slate-200 bg-slate-50 p-5">
          <div className="grid gap-4 lg:grid-cols-2">
            <label className="block text-sm">
              Zone
              <select
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.zoneId}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    zoneId: event.target.value,
                    pointOfInterestId: '',
                  }))
                }
              >
                <option value="">Select a zone</option>
                {zones.map((zone) => (
                  <option key={zone.id} value={zone.id}>
                    {zone.name}
                  </option>
                ))}
              </select>
            </label>

            <label className="block text-sm">
              Location Mode
              <select
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.locationMode}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    locationMode:
                      event.target.value === 'poi' ? 'poi' : 'coordinates',
                  }))
                }
              >
                <option value="coordinates">Coordinates</option>
                <option value="poi">Point of Interest</option>
              </select>
            </label>
          </div>

          {form.locationMode === 'poi' ? (
            <label className="block text-sm">
              Point of Interest
              <select
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.pointOfInterestId}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    pointOfInterestId: event.target.value,
                  }))
                }
              >
                <option value="">Select a point of interest</option>
                {pointsOfInterest.map((point) => (
                  <option key={point.id} value={point.id}>
                    {point.name}
                  </option>
                ))}
              </select>
              {selectedPointOfInterest ? (
                <p className="mt-2 text-xs text-slate-500">
                  Coordinates: {selectedPointOfInterest.lat},{' '}
                  {selectedPointOfInterest.lng}
                </p>
              ) : null}
            </label>
          ) : (
            <div className="grid gap-4 lg:grid-cols-2">
              <label className="block text-sm">
                <div className="flex items-center justify-between gap-3">
                  <span>Latitude</span>
                  <button
                    type="button"
                    className="rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 disabled:opacity-50"
                    onClick={handleUseCurrentLocation}
                    disabled={locationBusy}
                  >
                    {locationBusy ? 'Locating…' : 'Use Current Location'}
                  </button>
                </div>
                <input
                  className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                  value={form.latitude}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      latitude: event.target.value,
                    }))
                  }
                />
              </label>
              <label className="block text-sm">
                Longitude
                <input
                  className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                  value={form.longitude}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      longitude: event.target.value,
                    }))
                  }
                />
              </label>
            </div>
          )}

          <div className="grid gap-4 lg:grid-cols-2">
            <label className="block text-sm">
              Title
              <input
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.title}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, title: event.target.value }))
                }
              />
            </label>
            <label className="block text-sm">
              Reward Mode
              <select
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.rewardMode}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    rewardMode:
                      event.target.value === 'explicit' ? 'explicit' : 'random',
                  }))
                }
              >
                <option value="random">Random Reward Table</option>
                <option value="explicit">Explicit Rewards</option>
              </select>
            </label>
          </div>

          <label className="block text-sm">
            Description
            <textarea
              className="mt-1 min-h-[96px] w-full rounded-md border border-slate-300 bg-white p-2"
              value={form.description}
              onChange={(event) =>
                setForm((prev) => ({
                  ...prev,
                  description: event.target.value,
                }))
              }
            />
          </label>

          <DialogueMessageListEditor
            label="Dialogue"
            helperText="Each line in an exposition must have a speaking character."
            value={form.dialogue}
            onChange={(dialogue) => setForm((prev) => ({ ...prev, dialogue }))}
            characterOptions={characterOptions}
            requireCharacterSelection
          />

          <div className="grid gap-4 lg:grid-cols-2">
            <label className="block text-sm">
              Image URL
              <input
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.imageUrl}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, imageUrl: event.target.value }))
                }
              />
            </label>
            <label className="block text-sm">
              Thumbnail URL
              <input
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.thumbnailUrl}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    thumbnailUrl: event.target.value,
                  }))
                }
              />
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              type="button"
              className="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm text-slate-900 disabled:opacity-50"
              onClick={handleSaveTemplate}
              disabled={savingTemplate}
            >
              {savingTemplate ? 'Saving Template…' : 'Save As Template'}
            </button>
            <button
              type="button"
              className="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm text-slate-900 disabled:opacity-50"
              onClick={handleGenerateImage}
              disabled={!selectedId || saving}
            >
              Generate Image
            </button>
            {selectedId ? (
              <button
                type="button"
                className="rounded-md border border-red-300 bg-white px-4 py-2 text-sm text-red-700 disabled:opacity-50"
                onClick={handleDelete}
                disabled={saving}
              >
                Delete
              </button>
            ) : null}
          </div>

          {form.thumbnailUrl.trim() || form.imageUrl.trim() ? (
            <div className="rounded-xl border border-slate-200 bg-white p-3">
              <img
                src={form.thumbnailUrl.trim() || form.imageUrl.trim()}
                alt={form.title || 'Exposition preview'}
                className="max-h-64 rounded-lg object-cover"
              />
            </div>
          ) : null}

          {form.rewardMode === 'random' ? (
            <label className="block text-sm">
              Random Reward Size
              <select
                className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                value={form.randomRewardSize}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    randomRewardSize:
                      event.target.value === 'large'
                        ? 'large'
                        : event.target.value === 'medium'
                          ? 'medium'
                          : 'small',
                  }))
                }
              >
                <option value="small">Small</option>
                <option value="medium">Medium</option>
                <option value="large">Large</option>
              </select>
            </label>
          ) : (
            <div className="space-y-4">
              <div className="grid gap-4 lg:grid-cols-2">
                <label className="block text-sm">
                  Reward Experience
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                    type="number"
                    min={0}
                    value={form.rewardExperience}
                    onChange={(event) =>
                      setForm((prev) => ({
                        ...prev,
                        rewardExperience: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block text-sm">
                  Reward Gold
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2"
                    type="number"
                    min={0}
                    value={form.rewardGold}
                    onChange={(event) =>
                      setForm((prev) => ({
                        ...prev,
                        rewardGold: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <MaterialRewardsEditor
                value={form.materialRewards}
                onChange={(materialRewards) =>
                  setForm((prev) => ({ ...prev, materialRewards }))
                }
                title="Material Rewards"
              />

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Item Rewards</label>
                  <button
                    type="button"
                    className="rounded border border-slate-300 bg-white px-3 py-1 text-sm text-slate-900"
                    onClick={() =>
                      setForm((prev) => ({
                        ...prev,
                        itemRewards: [
                          ...prev.itemRewards,
                          { inventoryItemId: '', quantity: 1 },
                        ],
                      }))
                    }
                  >
                    Add Item
                  </button>
                </div>
                {form.itemRewards.length === 0 ? (
                  <div className="rounded border border-dashed border-slate-300 bg-white px-3 py-3 text-sm text-slate-500">
                    No item rewards configured.
                  </div>
                ) : (
                  form.itemRewards.map((reward, index) => (
                    <div
                      key={`item-${index}`}
                      className="grid gap-3 rounded border border-slate-300 bg-white p-3 md:grid-cols-[minmax(0,1fr)_120px_auto]"
                    >
                      <label className="block text-sm">
                        Item
                        <select
                          className="mt-1 w-full rounded border border-slate-300 bg-white p-2"
                          value={reward.inventoryItemId}
                          onChange={(event) =>
                            setForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, entryIndex) =>
                                  entryIndex === index
                                    ? {
                                        ...entry,
                                        inventoryItemId: event.target.value,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select an item</option>
                          {inventoryItems.map((item: InventoryItem) => (
                            <option key={item.id} value={item.id}>
                              {item.name}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="block text-sm">
                        Quantity
                        <input
                          className="mt-1 w-full rounded border border-slate-300 bg-white p-2"
                          type="number"
                          min={1}
                          step={1}
                          value={reward.quantity}
                          onChange={(event) =>
                            setForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, entryIndex) =>
                                  entryIndex === index
                                    ? {
                                        ...entry,
                                        quantity: parsePositiveInt(
                                          event.target.value,
                                          1
                                        ),
                                      }
                                    : entry
                              ),
                            }))
                          }
                        />
                      </label>
                      <div className="flex items-end">
                        <button
                          type="button"
                          className="rounded border border-red-300 bg-white px-3 py-2 text-sm text-red-700"
                          onClick={() =>
                            setForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.filter(
                                (_, entryIndex) => entryIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Spell Rewards</label>
                  <button
                    type="button"
                    className="rounded border border-slate-300 bg-white px-3 py-1 text-sm text-slate-900"
                    onClick={() =>
                      setForm((prev) => ({
                        ...prev,
                        spellRewards: [...prev.spellRewards, { spellId: '' }],
                      }))
                    }
                  >
                    Add Spell
                  </button>
                </div>
                {form.spellRewards.length === 0 ? (
                  <div className="rounded border border-dashed border-slate-300 bg-white px-3 py-3 text-sm text-slate-500">
                    No spell rewards configured.
                  </div>
                ) : (
                  form.spellRewards.map((reward, index) => (
                    <div
                      key={`spell-${index}`}
                      className="grid gap-3 rounded border border-slate-300 bg-white p-3 md:grid-cols-[minmax(0,1fr)_auto]"
                    >
                      <label className="block text-sm">
                        Spell
                        <select
                          className="mt-1 w-full rounded border border-slate-300 bg-white p-2"
                          value={reward.spellId}
                          onChange={(event) =>
                            setForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.map(
                                (entry, entryIndex) =>
                                  entryIndex === index
                                    ? { ...entry, spellId: event.target.value }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select a spell</option>
                          {spells.map((spell) => (
                            <option key={spell.id} value={spell.id}>
                              {spell.name}
                            </option>
                          ))}
                        </select>
                      </label>
                      <div className="flex items-end">
                        <button
                          type="button"
                          className="rounded border border-red-300 bg-white px-3 py-2 text-sm text-red-700"
                          onClick={() =>
                            setForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.filter(
                                (_, entryIndex) => entryIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

export default Expositions;
