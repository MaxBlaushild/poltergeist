import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { Spell } from '@poltergeist/types';

type ScenarioRewardItem = {
  inventoryItemId: number;
  quantity: number;
};

type ScenarioRewardSpell = {
  spellId: string;
};

type ScenarioFailurePenaltyMode = 'shared' | 'individual';
type ScenarioSuccessRewardMode = 'shared' | 'individual';
type ScenarioFailureDrainType = 'none' | 'flat' | 'percent';

type ScenarioFailureStatus = {
  name: string;
  description: string;
  effect: string;
  positive: boolean;
  durationSeconds: number;
  strengthMod: number;
  dexterityMod: number;
  constitutionMod: number;
  intelligenceMod: number;
  wisdomMod: number;
  charismaMod: number;
};

type ScenarioOption = {
  id?: string;
  optionText: string;
  successText: string;
  failureText: string;
  statTag: string;
  proficiencies: string[];
  difficulty?: number | null;
  rewardExperience: number;
  rewardGold: number;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: number;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: number;
  failureStatuses: ScenarioFailureStatus[];
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: number;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: number;
  successStatuses: ScenarioFailureStatus[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
};

type ScenarioRecord = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: number;
  rewardExperience: number;
  rewardGold: number;
  openEnded: boolean;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: number;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: number;
  failureStatuses: ScenarioFailureStatus[];
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: number;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: number;
  successStatuses: ScenarioFailureStatus[];
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
  attemptedByUser?: boolean;
};

type ScenarioFormState = {
  zoneId: string;
  latitude: string;
  longitude: string;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: string;
  openEnded: boolean;
  rewardExperience: string;
  rewardGold: string;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: string;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: string;
  failureStatuses: ScenarioFailureStatus[];
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: string;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: string;
  successStatuses: ScenarioFailureStatus[];
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
};

type ScenarioGenerationJob = {
  id: string;
  zoneId: string;
  status: string;
  openEnded: boolean;
  latitude?: number | null;
  longitude?: number | null;
  generatedScenarioId?: string | null;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ScenarioGenerationFormState = {
  zoneId: string;
  openEnded: boolean;
  includeLocation: boolean;
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

const defaultScenarioUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered scenario. Mysterious parchment sigil, subtle compass motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

const statTags = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
] as const;

const failureDrainTypes: ScenarioFailureDrainType[] = [
  'none',
  'flat',
  'percent',
];

const emptyFailureStatus = (): ScenarioFailureStatus => ({
  name: '',
  description: '',
  effect: '',
  positive: true,
  durationSeconds: 60,
  strengthMod: 0,
  dexterityMod: 0,
  constitutionMod: 0,
  intelligenceMod: 0,
  wisdomMod: 0,
  charismaMod: 0,
});

const emptyOption = (): ScenarioOption => ({
  optionText: '',
  successText: 'Your approach works, and momentum turns in your favor.',
  failureText: 'The attempt falls short, and the moment slips away.',
  statTag: 'charisma',
  proficiencies: [],
  difficulty: null,
  rewardExperience: 0,
  rewardGold: 0,
  failureHealthDrainType: 'none',
  failureHealthDrainValue: 0,
  failureManaDrainType: 'none',
  failureManaDrainValue: 0,
  failureStatuses: [],
  successHealthRestoreType: 'none',
  successHealthRestoreValue: 0,
  successManaRestoreType: 'none',
  successManaRestoreValue: 0,
  successStatuses: [],
  itemRewards: [],
  spellRewards: [],
});

const emptyFormState = (): ScenarioFormState => ({
  zoneId: '',
  latitude: '',
  longitude: '',
  prompt: '',
  imageUrl: '',
  thumbnailUrl: '',
  difficulty: '24',
  openEnded: false,
  rewardExperience: '0',
  rewardGold: '0',
  failurePenaltyMode: 'shared',
  failureHealthDrainType: 'none',
  failureHealthDrainValue: '0',
  failureManaDrainType: 'none',
  failureManaDrainValue: '0',
  failureStatuses: [],
  successRewardMode: 'shared',
  successHealthRestoreType: 'none',
  successHealthRestoreValue: '0',
  successManaRestoreType: 'none',
  successManaRestoreValue: '0',
  successStatuses: [],
  options: [emptyOption()],
  itemRewards: [],
  spellRewards: [],
});

const emptyGenerationFormState = (): ScenarioGenerationFormState => ({
  zoneId: '',
  openEnded: false,
  includeLocation: false,
  latitude: '',
  longitude: '',
});

const parseIntValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseCsv = (value: string): string[] => {
  return value
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);
};

const normalizeFailureDrainType = (
  value?: string | null
): ScenarioFailureDrainType => {
  if (value === 'flat' || value === 'percent') return value;
  return 'none';
};

const normalizeFailurePenaltyMode = (
  value?: string | null
): ScenarioFailurePenaltyMode => {
  if (value === 'individual') return 'individual';
  return 'shared';
};

const normalizeSuccessRewardMode = (
  value?: string | null
): ScenarioSuccessRewardMode => {
  if (value === 'individual') return 'individual';
  return 'shared';
};

const normalizeFailureStatus = (
  status?: Partial<ScenarioFailureStatus> | null
): ScenarioFailureStatus => {
  const base = emptyFailureStatus();
  if (!status) return base;
  return {
    ...base,
    ...status,
    name: (status.name ?? '').trim(),
    description: (status.description ?? '').trim(),
    effect: (status.effect ?? '').trim(),
    durationSeconds: Number.isFinite(status.durationSeconds)
      ? Number(status.durationSeconds)
      : base.durationSeconds,
    strengthMod: Number.isFinite(status.strengthMod)
      ? Number(status.strengthMod)
      : 0,
    dexterityMod: Number.isFinite(status.dexterityMod)
      ? Number(status.dexterityMod)
      : 0,
    constitutionMod: Number.isFinite(status.constitutionMod)
      ? Number(status.constitutionMod)
      : 0,
    intelligenceMod: Number.isFinite(status.intelligenceMod)
      ? Number(status.intelligenceMod)
      : 0,
    wisdomMod: Number.isFinite(status.wisdomMod) ? Number(status.wisdomMod) : 0,
    charismaMod: Number.isFinite(status.charismaMod)
      ? Number(status.charismaMod)
      : 0,
    positive: status.positive ?? true,
  };
};

const scenarioGenerationStatusBadgeClass = (status: string): string => {
  switch (status) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    default:
      return 'bg-gray-600';
  }
};

const formatDate = (value?: string): string => {
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

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Scenarios = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();
  const [spells, setSpells] = useState<Spell[]>([]);

  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState<ScenarioRecord[]>([]);
  const [query, setQuery] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [generationForm, setGenerationForm] =
    useState<ScenarioGenerationFormState>(emptyGenerationFormState);
  const [generationJobs, setGenerationJobs] = useState<ScenarioGenerationJob[]>(
    []
  );
  const [generationJobsLoading, setGenerationJobsLoading] = useState(false);
  const [generationSubmitting, setGenerationSubmitting] = useState(false);
  const [generationError, setGenerationError] = useState<string | null>(null);
  const [generationGeoLoading, setGenerationGeoLoading] = useState(false);
  const [scenarioUndiscoveredBusy, setScenarioUndiscoveredBusy] =
    useState(false);
  const [
    scenarioUndiscoveredStatusLoading,
    setScenarioUndiscoveredStatusLoading,
  ] = useState(false);
  const [scenarioUndiscoveredError, setScenarioUndiscoveredError] = useState<
    string | null
  >(null);
  const [scenarioUndiscoveredMessage, setScenarioUndiscoveredMessage] =
    useState<string | null>(null);
  const [scenarioUndiscoveredUrl, setScenarioUndiscoveredUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png'
  );
  const [scenarioUndiscoveredStatus, setScenarioUndiscoveredStatus] =
    useState('unknown');
  const [scenarioUndiscoveredExists, setScenarioUndiscoveredExists] =
    useState(false);
  const [scenarioUndiscoveredRequestedAt, setScenarioUndiscoveredRequestedAt] =
    useState<string | null>(null);
  const [
    scenarioUndiscoveredLastModified,
    setScenarioUndiscoveredLastModified,
  ] = useState<string | null>(null);
  const [
    scenarioUndiscoveredPreviewNonce,
    setScenarioUndiscoveredPreviewNonce,
  ] = useState<number>(Date.now());
  const [
    scenarioUndiscoveredPrompt,
    setScenarioUndiscoveredPrompt,
  ] = useState(defaultScenarioUndiscoveredIconPrompt);

  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ScenarioFormState>(emptyFormState);
  const [generatingScenarioId, setGeneratingScenarioId] = useState<
    string | null
  >(null);
  const [geoLoading, setGeoLoading] = useState(false);
  const [expandedScenarioImage, setExpandedScenarioImage] = useState<{
    url: string;
    title: string;
  } | null>(null);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(form.latitude);
  const formLongitudeRef = React.useRef(form.longitude);
  const generationMapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const generationMapRef = React.useRef<mapboxgl.Map | null>(null);
  const generationMarkerRef = React.useRef<mapboxgl.Marker | null>(null);
  const generationLatitudeRef = React.useRef(generationForm.latitude);
  const generationLongitudeRef = React.useRef(generationForm.longitude);
  const seenCompletedGenerationJobsRef = React.useRef<Set<string>>(new Set());

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response =
        await apiClient.get<ScenarioRecord[]>('/sonar/scenarios');
      setRecords(response);
    } catch (err) {
      console.error('Error loading scenarios:', err);
      setError('Failed to load scenarios.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  const refreshScenarioById = useCallback(
    async (scenarioId: string) => {
      const latest = await apiClient.get<ScenarioRecord>(
        `/sonar/scenarios/${scenarioId}`
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === scenarioId ? latest : record))
      );
      return latest;
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    let active = true;
    const loadSpells = async () => {
      try {
        const response = await apiClient.get<Spell[]>('/sonar/spells');
        if (!active) return;
        setSpells(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Error loading spells:', err);
        if (active) setSpells([]);
      }
    };
    void loadSpells();
    return () => {
      active = false;
    };
  }, [apiClient]);

  const loadGenerationJobs = useCallback(async () => {
    try {
      setGenerationJobsLoading(true);
      const response = await apiClient.get<ScenarioGenerationJob[]>(
        '/sonar/admin/scenario-generation-jobs',
        {
          limit: 25,
        }
      );
      setGenerationJobs(response);
      setGenerationError(null);
    } catch (err) {
      console.error('Error loading scenario generation jobs:', err);
      setGenerationError('Failed to load scenario generation jobs.');
    } finally {
      setGenerationJobsLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadGenerationJobs();
  }, [loadGenerationJobs]);

  useEffect(() => {
    if (!generationForm.zoneId && zones.length > 0) {
      setGenerationForm((prev) => ({ ...prev, zoneId: zones[0].id }));
    }
  }, [generationForm.zoneId, zones]);

  useEffect(() => {
    generationLatitudeRef.current = generationForm.latitude;
    generationLongitudeRef.current = generationForm.longitude;
  }, [generationForm.latitude, generationForm.longitude]);

  const hasActiveGenerationJobs = useMemo(
    () =>
      generationJobs.some(
        (job) => job.status === 'queued' || job.status === 'in_progress'
      ),
    [generationJobs]
  );

  useEffect(() => {
    if (!hasActiveGenerationJobs) return;
    const interval = window.setInterval(() => {
      void loadGenerationJobs();
    }, 3000);
    return () => {
      window.clearInterval(interval);
    };
  }, [hasActiveGenerationJobs, loadGenerationJobs]);

  useEffect(() => {
    const completedWithScenario = generationJobs.filter(
      (job) => job.status === 'completed' && !!job.generatedScenarioId
    );
    let shouldReloadScenarios = false;
    for (const job of completedWithScenario) {
      if (!seenCompletedGenerationJobsRef.current.has(job.id)) {
        seenCompletedGenerationJobsRef.current.add(job.id);
        shouldReloadScenarios = true;
      }
    }
    if (shouldReloadScenarios) {
      void load();
    }
  }, [generationJobs, load]);

  const setGenerationLocation = useCallback(
    (latitude: number, longitude: number) => {
      setGenerationForm((prev) => ({
        ...prev,
        includeLocation: true,
        latitude: latitude.toFixed(6),
        longitude: longitude.toFixed(6),
      }));
    },
    []
  );

  const handleQueueScenarioGeneration = async () => {
    if (!generationForm.zoneId) {
      setGenerationError('Please select a zone.');
      return;
    }
    const payload: {
      zoneId: string;
      openEnded: boolean;
      latitude?: number;
      longitude?: number;
    } = {
      zoneId: generationForm.zoneId,
      openEnded: generationForm.openEnded,
    };

    if (generationForm.includeLocation) {
      const latitude = parseFloatValue(generationForm.latitude, Number.NaN);
      const longitude = parseFloatValue(generationForm.longitude, Number.NaN);
      if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
        setGenerationError(
          'Location is enabled, so latitude and longitude are required.'
        );
        return;
      }
      payload.latitude = latitude;
      payload.longitude = longitude;
    }

    try {
      setGenerationSubmitting(true);
      setGenerationError(null);
      const created = await apiClient.post<ScenarioGenerationJob>(
        '/sonar/admin/scenario-generation-jobs',
        payload
      );
      setGenerationJobs((prev) => [created, ...prev]);
      setGenerationForm((prev) => ({
        ...prev,
        includeLocation: false,
        latitude: '',
        longitude: '',
      }));
    } catch (err) {
      console.error('Error queueing scenario generation job:', err);
      setGenerationError('Failed to queue scenario generation job.');
    } finally {
      setGenerationSubmitting(false);
    }
  };

  const handleUseCurrentGenerationLocation = useCallback(() => {
    if (!navigator.geolocation) {
      setGenerationError('Geolocation is not supported in this browser.');
      return;
    }
    setGenerationGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGenerationGeoLoading(false);
        setGenerationError(null);
        setGenerationLocation(
          position.coords.latitude,
          position.coords.longitude
        );
      },
      (geoError) => {
        setGenerationGeoLoading(false);
        setGenerationError(
          `Unable to get current location: ${geoError.message}`
        );
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setGenerationLocation]);

  const refreshUndiscoveredScenarioIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setScenarioUndiscoveredStatusLoading(true);
        setScenarioUndiscoveredError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/scenario-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setScenarioUndiscoveredUrl(url);
        }
        setScenarioUndiscoveredStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setScenarioUndiscoveredExists(Boolean(response?.exists));
        setScenarioUndiscoveredRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setScenarioUndiscoveredLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setScenarioUndiscoveredPreviewNonce(Date.now());
        if (showMessage) {
          setScenarioUndiscoveredMessage(
            'Undiscovered scenario icon status refreshed.'
          );
        }
      } catch (err) {
        console.error('Failed to load undiscovered scenario icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load undiscovered scenario icon status.';
        setScenarioUndiscoveredError(message);
      } finally {
        setScenarioUndiscoveredStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateUndiscoveredScenarioIcon = useCallback(async () => {
    const prompt = scenarioUndiscoveredPrompt.trim();
    if (!prompt) {
      setScenarioUndiscoveredError('Prompt is required.');
      return;
    }
    try {
      setScenarioUndiscoveredBusy(true);
      setScenarioUndiscoveredError(null);
      setScenarioUndiscoveredMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/scenario-undiscovered',
        { prompt }
      );
      setScenarioUndiscoveredMessage(
        'Undiscovered scenario icon queued for generation.'
      );
      await refreshUndiscoveredScenarioIconStatus();
    } catch (err) {
      console.error('Failed to generate undiscovered scenario icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate undiscovered scenario icon.';
      setScenarioUndiscoveredError(message);
    } finally {
      setScenarioUndiscoveredBusy(false);
    }
  }, [
    apiClient,
    refreshUndiscoveredScenarioIconStatus,
    scenarioUndiscoveredPrompt,
  ]);

  const handleDeleteUndiscoveredScenarioIcon = useCallback(async () => {
    try {
      setScenarioUndiscoveredBusy(true);
      setScenarioUndiscoveredError(null);
      setScenarioUndiscoveredMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/scenario-undiscovered'
      );
      setScenarioUndiscoveredMessage('Undiscovered scenario icon deleted.');
      await refreshUndiscoveredScenarioIconStatus();
    } catch (err) {
      console.error('Failed to delete undiscovered scenario icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete undiscovered scenario icon.';
      setScenarioUndiscoveredError(message);
    } finally {
      setScenarioUndiscoveredBusy(false);
    }
  }, [apiClient, refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    void refreshUndiscoveredScenarioIconStatus();
  }, [refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    if (
      scenarioUndiscoveredStatus !== 'queued' &&
      scenarioUndiscoveredStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshUndiscoveredScenarioIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [scenarioUndiscoveredStatus, refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    if (!generationForm.includeLocation) {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      generationMapRef.current?.remove();
      generationMapRef.current = null;
      return;
    }
    if (!generationMapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (generationMapRef.current) return;

    const parsedLat = Number.parseFloat(generationLatitudeRef.current);
    const parsedLng = Number.parseFloat(generationLongitudeRef.current);
    const selectedZone = zones.find(
      (zone) => zone.id === generationForm.zoneId
    );
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: generationMapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setGenerationLocation(event.lngLat.lat, event.lngLat.lng);
    });

    generationMapRef.current = map;

    return () => {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      map.remove();
      generationMapRef.current = null;
    };
  }, [
    generationForm.includeLocation,
    generationForm.zoneId,
    setGenerationLocation,
    zones,
  ]);

  useEffect(() => {
    if (!generationForm.includeLocation) return;
    if (!generationMapRef.current) return;

    const latitude = Number.parseFloat(generationForm.latitude);
    const longitude = Number.parseFloat(generationForm.longitude);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      return;
    }

    if (!generationMarkerRef.current) {
      generationMarkerRef.current = new mapboxgl.Marker({ color: '#2563eb' })
        .setLngLat([longitude, latitude])
        .addTo(generationMapRef.current);
    } else {
      generationMarkerRef.current.setLngLat([longitude, latitude]);
    }

    generationMapRef.current.easeTo({
      center: [longitude, latitude],
      duration: 350,
    });
  }, [
    generationForm.includeLocation,
    generationForm.latitude,
    generationForm.longitude,
  ]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return records;
    return records.filter((record) => {
      const zoneName =
        zones.find((zone) => zone.id === record.zoneId)?.name ?? '';
      return (
        record.prompt.toLowerCase().includes(q) ||
        zoneName.toLowerCase().includes(q) ||
        record.id.toLowerCase().includes(q)
      );
    });
  }, [query, records, zones]);

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyFormState());
    setShowModal(true);
  };

  const openEdit = (record: ScenarioRecord) => {
    setEditingId(record.id);
    setForm({
      zoneId: record.zoneId,
      latitude: record.latitude.toString(),
      longitude: record.longitude.toString(),
      prompt: record.prompt,
      imageUrl: record.imageUrl,
      thumbnailUrl: record.thumbnailUrl ?? '',
      difficulty: record.difficulty.toString(),
      openEnded: record.openEnded,
      rewardExperience: record.rewardExperience.toString(),
      rewardGold: record.rewardGold.toString(),
      failurePenaltyMode: normalizeFailurePenaltyMode(
        record.failurePenaltyMode
      ),
      failureHealthDrainType: normalizeFailureDrainType(
        record.failureHealthDrainType
      ),
      failureHealthDrainValue: String(record.failureHealthDrainValue ?? 0),
      failureManaDrainType: normalizeFailureDrainType(
        record.failureManaDrainType
      ),
      failureManaDrainValue: String(record.failureManaDrainValue ?? 0),
      failureStatuses: (record.failureStatuses ?? []).map((status) =>
        normalizeFailureStatus(status)
      ),
      successRewardMode: normalizeSuccessRewardMode(record.successRewardMode),
      successHealthRestoreType: normalizeFailureDrainType(
        record.successHealthRestoreType
      ),
      successHealthRestoreValue: String(record.successHealthRestoreValue ?? 0),
      successManaRestoreType: normalizeFailureDrainType(
        record.successManaRestoreType
      ),
      successManaRestoreValue: String(record.successManaRestoreValue ?? 0),
      successStatuses: (record.successStatuses ?? []).map((status) =>
        normalizeFailureStatus(status)
      ),
      options:
        record.options.length > 0
          ? record.options.map((option) => ({
              ...option,
              successText:
                option.successText?.trim() ||
                'Your approach works, and momentum turns in your favor.',
              failureText:
                option.failureText?.trim() ||
                'The attempt falls short, and the moment slips away.',
              failureHealthDrainType: normalizeFailureDrainType(
                option.failureHealthDrainType
              ),
              failureHealthDrainValue: option.failureHealthDrainValue ?? 0,
              failureManaDrainType: normalizeFailureDrainType(
                option.failureManaDrainType
              ),
              failureManaDrainValue: option.failureManaDrainValue ?? 0,
              failureStatuses: (option.failureStatuses ?? []).map((status) =>
                normalizeFailureStatus(status)
              ),
              successHealthRestoreType: normalizeFailureDrainType(
                option.successHealthRestoreType
              ),
              successHealthRestoreValue: option.successHealthRestoreValue ?? 0,
              successManaRestoreType: normalizeFailureDrainType(
                option.successManaRestoreType
              ),
              successManaRestoreValue: option.successManaRestoreValue ?? 0,
              successStatuses: (option.successStatuses ?? []).map((status) =>
                normalizeFailureStatus(status)
              ),
              spellRewards: (option.spellRewards ?? []).map((reward) => ({
                spellId: reward.spellId,
              })),
            }))
          : [emptyOption()],
      itemRewards: record.itemRewards,
      spellRewards: (record.spellRewards ?? []).map((reward) => ({
        spellId: reward.spellId,
      })),
    });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingId(null);
    setForm(emptyFormState());
  };

  useEffect(() => {
    formLatitudeRef.current = form.latitude;
    formLongitudeRef.current = form.longitude;
  }, [form.latitude, form.longitude]);

  const setFormLocation = useCallback((latitude: number, longitude: number) => {
    setForm((prev) => ({
      ...prev,
      latitude: latitude.toFixed(6),
      longitude: longitude.toFixed(6),
    }));
  }, []);

  const formPayload = () => {
    const toStatusPayload = (statuses: ScenarioFailureStatus[]) =>
      statuses
        .map((status) => ({
          ...status,
          name: status.name.trim(),
          description: status.description.trim(),
          effect: status.effect.trim(),
          durationSeconds: parseIntValue(String(status.durationSeconds), 0),
        }))
        .filter((status) => status.name !== '' && status.durationSeconds > 0);

    const scenarioPenaltyMode: ScenarioFailurePenaltyMode = form.openEnded
      ? 'shared'
      : form.failurePenaltyMode;
    const useOptionPenalties =
      !form.openEnded && scenarioPenaltyMode === 'individual';
    const successRewardMode: ScenarioSuccessRewardMode = form.openEnded
      ? 'shared'
      : form.successRewardMode;
    const useOptionSuccessRewards =
      !form.openEnded && successRewardMode === 'individual';

    return {
      zoneId: form.zoneId,
      latitude: parseFloatValue(form.latitude),
      longitude: parseFloatValue(form.longitude),
      prompt: form.prompt.trim(),
      imageUrl: form.imageUrl.trim(),
      thumbnailUrl: form.thumbnailUrl.trim(),
      difficulty: parseIntValue(form.difficulty, 24),
      openEnded: form.openEnded,
      rewardExperience: form.openEnded
        ? parseIntValue(form.rewardExperience)
        : 0,
      rewardGold: form.openEnded ? parseIntValue(form.rewardGold) : 0,
      failurePenaltyMode: scenarioPenaltyMode,
      successRewardMode,
      failureHealthDrainType: form.failureHealthDrainType,
      failureHealthDrainValue:
        form.failureHealthDrainType === 'none'
          ? 0
          : parseIntValue(form.failureHealthDrainValue, 0),
      failureManaDrainType: form.failureManaDrainType,
      failureManaDrainValue:
        form.failureManaDrainType === 'none'
          ? 0
          : parseIntValue(form.failureManaDrainValue, 0),
      failureStatuses: toStatusPayload(form.failureStatuses),
      successHealthRestoreType: form.successHealthRestoreType,
      successHealthRestoreValue:
        form.successHealthRestoreType === 'none'
          ? 0
          : parseIntValue(form.successHealthRestoreValue, 0),
      successManaRestoreType: form.successManaRestoreType,
      successManaRestoreValue:
        form.successManaRestoreType === 'none'
          ? 0
          : parseIntValue(form.successManaRestoreValue, 0),
      successStatuses: toStatusPayload(form.successStatuses),
      options: form.openEnded
        ? []
        : form.options.map((option) => ({
            optionText: option.optionText.trim(),
            successText: option.successText.trim(),
            failureText: option.failureText.trim(),
            statTag: option.statTag,
            proficiencies: option.proficiencies,
            difficulty: option.difficulty,
            rewardExperience: option.rewardExperience,
            rewardGold: option.rewardGold,
            failureHealthDrainType: useOptionPenalties
              ? option.failureHealthDrainType
              : 'none',
            failureHealthDrainValue:
              useOptionPenalties && option.failureHealthDrainType !== 'none'
                ? option.failureHealthDrainValue
                : 0,
            failureManaDrainType: useOptionPenalties
              ? option.failureManaDrainType
              : 'none',
            failureManaDrainValue:
              useOptionPenalties && option.failureManaDrainType !== 'none'
                ? option.failureManaDrainValue
                : 0,
            failureStatuses: useOptionPenalties
              ? toStatusPayload(option.failureStatuses)
              : [],
            successHealthRestoreType: useOptionSuccessRewards
              ? option.successHealthRestoreType
              : 'none',
            successHealthRestoreValue:
              useOptionSuccessRewards &&
              option.successHealthRestoreType !== 'none'
                ? option.successHealthRestoreValue
                : 0,
            successManaRestoreType: useOptionSuccessRewards
              ? option.successManaRestoreType
              : 'none',
            successManaRestoreValue:
              useOptionSuccessRewards &&
              option.successManaRestoreType !== 'none'
                ? option.successManaRestoreValue
                : 0,
            successStatuses: useOptionSuccessRewards
              ? toStatusPayload(option.successStatuses)
              : [],
            itemRewards: option.itemRewards,
            spellRewards: option.spellRewards
              .map((reward) => ({ spellId: reward.spellId.trim() }))
              .filter((reward) => reward.spellId.length > 0),
          })),
      itemRewards: form.openEnded ? form.itemRewards : [],
      spellRewards: form.openEnded
        ? form.spellRewards
            .map((reward) => ({ spellId: reward.spellId.trim() }))
            .filter((reward) => reward.spellId.length > 0)
        : [],
    };
  };

  const save = async () => {
    try {
      const payload = formPayload();
      if (
        !payload.zoneId ||
        !payload.prompt ||
        !payload.imageUrl ||
        !payload.thumbnailUrl
      ) {
        alert('Zone, prompt, image URL, and thumbnail URL are required.');
        return;
      }
      if (payload.openEnded === false && payload.options.length === 0) {
        alert('Non-open-ended scenarios need at least one option.');
        return;
      }

      if (editingId) {
        const updated = await apiClient.put<ScenarioRecord>(
          `/sonar/scenarios/${editingId}`,
          payload
        );
        setRecords((prev) =>
          prev.map((record) => (record.id === updated.id ? updated : record))
        );
      } else {
        const created = await apiClient.post<ScenarioRecord>(
          '/sonar/scenarios',
          payload
        );
        setRecords((prev) => [created, ...prev]);
      }
      closeModal();
    } catch (err) {
      console.error('Error saving scenario:', err);
      alert('Failed to save scenario. Check required fields and try again.');
    }
  };

  const confirmDelete = async () => {
    if (!deleteId) return;
    try {
      await apiClient.delete(`/sonar/scenarios/${deleteId}`);
      setRecords((prev) => prev.filter((record) => record.id !== deleteId));
      setDeleteId(null);
    } catch (err) {
      console.error('Error deleting scenario:', err);
      alert('Failed to delete scenario.');
    }
  };

  const handleGenerateScenarioImage = async (record: ScenarioRecord) => {
    if (generatingScenarioId) return;
    setGeneratingScenarioId(record.id);
    const previousImageURL = (record.imageUrl || '').trim();
    try {
      await apiClient.post(`/sonar/scenarios/${record.id}/generate-image`, {});

      for (let attempt = 0; attempt < 18; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 1200));
        const latest = await refreshScenarioById(record.id);
        const nextImageURL = (latest.imageUrl || '').trim();
        if (nextImageURL && nextImageURL !== previousImageURL) {
          break;
        }
      }
    } catch (err) {
      console.error('Error generating scenario image:', err);
      alert('Failed to queue scenario image generation.');
    } finally {
      setGeneratingScenarioId(null);
    }
  };

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setFormLocation(position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${error.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setFormLocation]);

  useEffect(() => {
    if (!showModal) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === form.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setFormLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.remove();
      mapRef.current = null;
    };
  }, [form.zoneId, setFormLocation, showModal, zones]);

  useEffect(() => {
    if (!showModal) return;
    if (!mapRef.current) return;

    const lat = Number.parseFloat(form.latitude);
    const lng = Number.parseFloat(form.longitude);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      markerRef.current?.remove();
      markerRef.current = null;
      return;
    }

    if (!markerRef.current) {
      markerRef.current = new mapboxgl.Marker({ color: '#dc2626' })
        .setLngLat([lng, lat])
        .addTo(mapRef.current);
    } else {
      markerRef.current.setLngLat([lng, lat]);
    }

    mapRef.current.easeTo({ center: [lng, lat], duration: 350 });
  }, [form.latitude, form.longitude, showModal]);

  const updateOption = (index: number, next: Partial<ScenarioOption>) => {
    setForm((prev) => {
      const options = [...prev.options];
      options[index] = { ...options[index], ...next };
      return { ...prev, options };
    });
  };

  const addOption = () => {
    setForm((prev) => ({ ...prev, options: [...prev.options, emptyOption()] }));
  };

  const removeOption = (index: number) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.filter((_, i) => i !== index),
    }));
  };

  const addOptionItemReward = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = [
        ...option.itemRewards,
        { inventoryItemId: 0, quantity: 1 },
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionItemReward = (
    optionIndex: number,
    rewardIndex: number,
    next: Partial<ScenarioRewardItem>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const rewards = [...option.itemRewards];
      rewards[rewardIndex] = { ...rewards[rewardIndex], ...next };
      option.itemRewards = rewards;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionItemReward = (optionIndex: number, rewardIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = option.itemRewards.filter(
        (_, i) => i !== rewardIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addOptionSpellReward = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.spellRewards = [...option.spellRewards, { spellId: '' }];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionSpellReward = (
    optionIndex: number,
    rewardIndex: number,
    next: Partial<ScenarioRewardSpell>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const rewards = [...option.spellRewards];
      rewards[rewardIndex] = { ...rewards[rewardIndex], ...next };
      option.spellRewards = rewards;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionSpellReward = (
    optionIndex: number,
    rewardIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.spellRewards = option.spellRewards.filter(
        (_, i) => i !== rewardIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addScenarioItemReward = () => {
    setForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { inventoryItemId: 0, quantity: 1 }],
    }));
  };

  const updateScenarioItemReward = (
    index: number,
    next: Partial<ScenarioRewardItem>
  ) => {
    setForm((prev) => {
      const rewards = [...prev.itemRewards];
      rewards[index] = { ...rewards[index], ...next };
      return { ...prev, itemRewards: rewards };
    });
  };

  const removeScenarioItemReward = (index: number) => {
    setForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.filter((_, i) => i !== index),
    }));
  };

  const addScenarioSpellReward = () => {
    setForm((prev) => ({
      ...prev,
      spellRewards: [...prev.spellRewards, { spellId: '' }],
    }));
  };

  const updateScenarioSpellReward = (
    index: number,
    next: Partial<ScenarioRewardSpell>
  ) => {
    setForm((prev) => {
      const rewards = [...prev.spellRewards];
      rewards[index] = { ...rewards[index], ...next };
      return { ...prev, spellRewards: rewards };
    });
  };

  const removeScenarioSpellReward = (index: number) => {
    setForm((prev) => ({
      ...prev,
      spellRewards: prev.spellRewards.filter((_, i) => i !== index),
    }));
  };

  const addScenarioFailureStatus = () => {
    setForm((prev) => ({
      ...prev,
      failureStatuses: [...prev.failureStatuses, emptyFailureStatus()],
    }));
  };

  const updateScenarioFailureStatus = (
    index: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const statuses = [...prev.failureStatuses];
      statuses[index] = { ...statuses[index], ...next };
      return { ...prev, failureStatuses: statuses };
    });
  };

  const removeScenarioFailureStatus = (index: number) => {
    setForm((prev) => ({
      ...prev,
      failureStatuses: prev.failureStatuses.filter((_, i) => i !== index),
    }));
  };

  const addOptionFailureStatus = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.failureStatuses = [
        ...option.failureStatuses,
        emptyFailureStatus(),
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionFailureStatus = (
    optionIndex: number,
    statusIndex: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const statuses = [...option.failureStatuses];
      statuses[statusIndex] = { ...statuses[statusIndex], ...next };
      option.failureStatuses = statuses;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionFailureStatus = (
    optionIndex: number,
    statusIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.failureStatuses = option.failureStatuses.filter(
        (_, i) => i !== statusIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addScenarioSuccessStatus = () => {
    setForm((prev) => ({
      ...prev,
      successStatuses: [...prev.successStatuses, emptyFailureStatus()],
    }));
  };

  const updateScenarioSuccessStatus = (
    index: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const statuses = [...prev.successStatuses];
      statuses[index] = { ...statuses[index], ...next };
      return { ...prev, successStatuses: statuses };
    });
  };

  const removeScenarioSuccessStatus = (index: number) => {
    setForm((prev) => ({
      ...prev,
      successStatuses: prev.successStatuses.filter((_, i) => i !== index),
    }));
  };

  const addOptionSuccessStatus = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.successStatuses = [
        ...option.successStatuses,
        emptyFailureStatus(),
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionSuccessStatus = (
    optionIndex: number,
    statusIndex: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const statuses = [...option.successStatuses];
      statuses[statusIndex] = { ...statuses[statusIndex], ...next };
      option.successStatuses = statuses;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionSuccessStatus = (
    optionIndex: number,
    statusIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.successStatuses = option.successStatuses.filter(
        (_, i) => i !== statusIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const renderFailurePenaltyEditor = (config: {
    title: string;
    healthDrainType: ScenarioFailureDrainType;
    healthDrainValue: number;
    manaDrainType: ScenarioFailureDrainType;
    manaDrainValue: number;
    statuses: ScenarioFailureStatus[];
    onHealthDrainTypeChange: (value: ScenarioFailureDrainType) => void;
    onHealthDrainValueChange: (value: number) => void;
    onManaDrainTypeChange: (value: ScenarioFailureDrainType) => void;
    onManaDrainValueChange: (value: number) => void;
    onAddStatus: () => void;
    onUpdateStatus: (
      index: number,
      next: Partial<ScenarioFailureStatus>
    ) => void;
    onRemoveStatus: (index: number) => void;
  }) => (
    <div className="border rounded-md p-3 mt-3">
      <div className="font-medium mb-2">{config.title}</div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
        <label className="text-sm">
          Health Drain Type
          <select
            value={config.healthDrainType}
            onChange={(e) =>
              config.onHealthDrainTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Health Drain Value
          <input
            value={config.healthDrainValue}
            onChange={(e) =>
              config.onHealthDrainValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.healthDrainType === 'percent' ? 100 : undefined}
            disabled={config.healthDrainType === 'none'}
          />
        </label>
        <label className="text-sm">
          Mana Drain Type
          <select
            value={config.manaDrainType}
            onChange={(e) =>
              config.onManaDrainTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Mana Drain Value
          <input
            value={config.manaDrainValue}
            onChange={(e) =>
              config.onManaDrainValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.manaDrainType === 'percent' ? 100 : undefined}
            disabled={config.manaDrainType === 'none'}
          />
        </label>
      </div>

      <div className="flex items-center justify-between mb-2">
        <div className="font-medium text-sm">Failure Statuses</div>
        <button
          type="button"
          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
          onClick={config.onAddStatus}
        >
          Add Status
        </button>
      </div>

      {config.statuses.length === 0 && (
        <div className="text-sm text-gray-600">
          No failure statuses configured.
        </div>
      )}

      {config.statuses.map((status, statusIndex) => (
        <div key={statusIndex} className="border rounded-md p-3 mb-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2">
            <label className="text-sm">
              Name
              <input
                value={status.name}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { name: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm">
              Duration (seconds)
              <input
                value={status.durationSeconds}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    durationSeconds: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
                min={1}
              />
            </label>
            <label className="text-sm md:col-span-2">
              Description
              <input
                value={status.description}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    description: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm md:col-span-2">
              Effect
              <input
                value={status.effect}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { effect: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
          </div>

          <label className="inline-flex items-center gap-2 text-sm mb-2">
            <input
              type="checkbox"
              checked={status.positive}
              onChange={(e) =>
                config.onUpdateStatus(statusIndex, {
                  positive: e.target.checked,
                })
              }
            />
            Positive status
          </label>

          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 mb-2">
            <label className="text-xs">
              STR
              <input
                value={status.strengthMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    strengthMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              DEX
              <input
                value={status.dexterityMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    dexterityMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CON
              <input
                value={status.constitutionMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    constitutionMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              INT
              <input
                value={status.intelligenceMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    intelligenceMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              WIS
              <input
                value={status.wisdomMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    wisdomMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CHA
              <input
                value={status.charismaMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    charismaMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
          </div>

          <button
            type="button"
            className="bg-red-500 text-white px-3 py-1 rounded-md text-xs"
            onClick={() => config.onRemoveStatus(statusIndex)}
          >
            Remove Status
          </button>
        </div>
      ))}
    </div>
  );

  const renderSuccessRewardEditor = (config: {
    title: string;
    healthRestoreType: ScenarioFailureDrainType;
    healthRestoreValue: number;
    manaRestoreType: ScenarioFailureDrainType;
    manaRestoreValue: number;
    statuses: ScenarioFailureStatus[];
    onHealthRestoreTypeChange: (value: ScenarioFailureDrainType) => void;
    onHealthRestoreValueChange: (value: number) => void;
    onManaRestoreTypeChange: (value: ScenarioFailureDrainType) => void;
    onManaRestoreValueChange: (value: number) => void;
    onAddStatus: () => void;
    onUpdateStatus: (
      index: number,
      next: Partial<ScenarioFailureStatus>
    ) => void;
    onRemoveStatus: (index: number) => void;
  }) => (
    <div className="border rounded-md p-3 mt-3">
      <div className="font-medium mb-2">{config.title}</div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
        <label className="text-sm">
          Health Restore Type
          <select
            value={config.healthRestoreType}
            onChange={(e) =>
              config.onHealthRestoreTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Health Restore Value
          <input
            value={config.healthRestoreValue}
            onChange={(e) =>
              config.onHealthRestoreValueChange(
                parseIntValue(e.target.value, 0)
              )
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.healthRestoreType === 'percent' ? 100 : undefined}
            disabled={config.healthRestoreType === 'none'}
          />
        </label>
        <label className="text-sm">
          Mana Restore Type
          <select
            value={config.manaRestoreType}
            onChange={(e) =>
              config.onManaRestoreTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Mana Restore Value
          <input
            value={config.manaRestoreValue}
            onChange={(e) =>
              config.onManaRestoreValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.manaRestoreType === 'percent' ? 100 : undefined}
            disabled={config.manaRestoreType === 'none'}
          />
        </label>
      </div>

      <div className="flex items-center justify-between mb-2">
        <div className="font-medium text-sm">Success Statuses</div>
        <button
          type="button"
          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
          onClick={config.onAddStatus}
        >
          Add Status
        </button>
      </div>

      {config.statuses.length === 0 && (
        <div className="text-sm text-gray-600">
          No success statuses configured.
        </div>
      )}

      {config.statuses.map((status, statusIndex) => (
        <div key={statusIndex} className="border rounded-md p-3 mb-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2">
            <label className="text-sm">
              Name
              <input
                value={status.name}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { name: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm">
              Duration (seconds)
              <input
                value={status.durationSeconds}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    durationSeconds: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
                min={1}
              />
            </label>
            <label className="text-sm md:col-span-2">
              Description
              <input
                value={status.description}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    description: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm md:col-span-2">
              Effect
              <input
                value={status.effect}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { effect: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
          </div>

          <label className="inline-flex items-center gap-2 text-sm mb-2">
            <input
              type="checkbox"
              checked={status.positive}
              onChange={(e) =>
                config.onUpdateStatus(statusIndex, {
                  positive: e.target.checked,
                })
              }
            />
            Positive status
          </label>

          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 mb-2">
            <label className="text-xs">
              STR
              <input
                value={status.strengthMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    strengthMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              DEX
              <input
                value={status.dexterityMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    dexterityMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CON
              <input
                value={status.constitutionMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    constitutionMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              INT
              <input
                value={status.intelligenceMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    intelligenceMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              WIS
              <input
                value={status.wisdomMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    wisdomMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CHA
              <input
                value={status.charismaMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    charismaMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
          </div>

          <button
            type="button"
            className="bg-red-500 text-white px-3 py-1 rounded-md text-xs"
            onClick={() => config.onRemoveStatus(statusIndex)}
          >
            Remove Status
          </button>
        </div>
      ))}
    </div>
  );

  if (loading) {
    return <div className="m-10">Loading scenarios...</div>;
  }

  return (
    <div className="m-10">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Scenarios</h1>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={openCreate}
        >
          Create Scenario
        </button>
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">Undiscovered Scenario Icon</h2>
          <div className="flex gap-2">
            <button
              type="button"
              className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={() => void refreshUndiscoveredScenarioIconStatus(true)}
              disabled={scenarioUndiscoveredStatusLoading}
            >
              {scenarioUndiscoveredStatusLoading
                ? 'Refreshing…'
                : 'Refresh Status'}
            </button>
            <button
              type="button"
              className="bg-indigo-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleGenerateUndiscoveredScenarioIcon}
              disabled={
                scenarioUndiscoveredBusy || scenarioUndiscoveredStatusLoading
              }
            >
              {scenarioUndiscoveredBusy ? 'Working…' : 'Generate Icon'}
            </button>
            <button
              type="button"
              className="bg-red-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleDeleteUndiscoveredScenarioIcon}
              disabled={
                scenarioUndiscoveredBusy || scenarioUndiscoveredStatusLoading
              }
            >
              {scenarioUndiscoveredBusy ? 'Working…' : 'Delete Icon'}
            </button>
          </div>
        </div>
        <div className="mb-2">
          <span
            className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
              scenarioUndiscoveredStatus
            )}`}
          >
            {scenarioUndiscoveredStatus || 'unknown'}
          </span>
        </div>
        <div className="text-xs text-gray-600 break-all">
          URL: {scenarioUndiscoveredUrl}
        </div>
        <div className="text-xs text-gray-600 mt-1">
          Requested: {formatDate(scenarioUndiscoveredRequestedAt ?? undefined)}
          {' · '}
          Last updated:{' '}
          {formatDate(scenarioUndiscoveredLastModified ?? undefined)}
        </div>
        <label className="block text-sm mt-3">
          Generation Prompt
          <textarea
            className="w-full border rounded-md p-2 mt-1 min-h-[88px]"
            value={scenarioUndiscoveredPrompt}
            onChange={(event) =>
              setScenarioUndiscoveredPrompt(event.target.value)
            }
            placeholder="Prompt used to generate the undiscovered scenario icon."
          />
        </label>
        {scenarioUndiscoveredExists ? (
          <div className="mt-3">
            <img
              src={`${scenarioUndiscoveredUrl}?v=${scenarioUndiscoveredPreviewNonce}`}
              alt="Undiscovered scenario icon preview"
              className="w-24 h-24 object-cover border rounded-md bg-gray-50"
            />
          </div>
        ) : (
          <div className="text-xs text-gray-500 mt-2">
            No icon currently found at this URL.
          </div>
        )}
        {scenarioUndiscoveredMessage ? (
          <div className="text-sm text-emerald-700 mt-2">
            {scenarioUndiscoveredMessage}
          </div>
        ) : null}
        {scenarioUndiscoveredError ? (
          <div className="text-sm text-red-600 mt-2">
            {scenarioUndiscoveredError}
          </div>
        ) : null}
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">Generate Scenario (Async)</h2>
          <button
            type="button"
            className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
            onClick={() => void loadGenerationJobs()}
            disabled={generationJobsLoading}
          >
            {generationJobsLoading ? 'Refreshing…' : 'Refresh Jobs'}
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
          <label className="text-sm">
            Zone
            <select
              value={generationForm.zoneId}
              onChange={(e) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  zoneId: e.target.value,
                }))
              }
              className="w-full border rounded-md p-2"
            >
              <option value="">Select zone</option>
              {zones.map((zone) => (
                <option key={zone.id} value={zone.id}>
                  {zone.name}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm">
            Scenario Type
            <select
              value={generationForm.openEnded ? 'open_ended' : 'choice'}
              onChange={(e) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  openEnded: e.target.value === 'open_ended',
                }))
              }
              className="w-full border rounded-md p-2"
            >
              <option value="choice">Choice (Options)</option>
              <option value="open_ended">Open-Ended (Free Text)</option>
            </select>
          </label>
        </div>

        <label className="inline-flex items-center gap-2 mb-3 text-sm">
          <input
            type="checkbox"
            checked={generationForm.includeLocation}
            onChange={(e) =>
              setGenerationForm((prev) => ({
                ...prev,
                includeLocation: e.target.checked,
                latitude: e.target.checked ? prev.latitude : '',
                longitude: e.target.checked ? prev.longitude : '',
              }))
            }
          />
          Provide a specific location (optional)
        </label>

        {generationForm.includeLocation && (
          <div className="border rounded-md p-3 mb-3">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
              <label className="text-sm">
                Latitude
                <input
                  value={generationForm.latitude}
                  onChange={(e) =>
                    setGenerationForm((prev) => ({
                      ...prev,
                      latitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <label className="text-sm">
                Longitude
                <input
                  value={generationForm.longitude}
                  onChange={(e) =>
                    setGenerationForm((prev) => ({
                      ...prev,
                      longitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
            </div>
            <div className="mb-2">
              <button
                type="button"
                className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                onClick={handleUseCurrentGenerationLocation}
                disabled={generationGeoLoading}
              >
                {generationGeoLoading
                  ? 'Locating…'
                  : 'Use Current Browser Location'}
              </button>
            </div>
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">Map Location Picker</span>
                <span className="text-xs text-gray-500">
                  Click map to set latitude/longitude
                </span>
              </div>
              {mapboxgl.accessToken ? (
                <div
                  ref={generationMapContainerRef}
                  className="w-full h-56 border rounded-md"
                />
              ) : (
                <div className="w-full border rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                  Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                  unavailable.
                </div>
              )}
            </div>
          </div>
        )}

        <div className="flex flex-wrap items-center gap-2 mb-3">
          <button
            type="button"
            className="bg-indigo-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
            onClick={handleQueueScenarioGeneration}
            disabled={generationSubmitting}
          >
            {generationSubmitting ? 'Queueing…' : 'Queue Scenario Generation'}
          </button>
        </div>

        {generationError && (
          <div className="mb-3 text-red-600 text-sm">{generationError}</div>
        )}

        <div className="font-medium mb-2">Recent Generation Jobs</div>
        {generationJobsLoading && generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">Loading jobs...</div>
        ) : generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">No jobs yet.</div>
        ) : (
          <div className="grid gap-2">
            {generationJobs.map((job) => {
              const zoneName =
                zones.find((zone) => zone.id === job.zoneId)?.name ??
                job.zoneId;
              const record = job.generatedScenarioId
                ? records.find(
                    (scenario) => scenario.id === job.generatedScenarioId
                  )
                : undefined;
              return (
                <div
                  key={job.id}
                  className="border rounded-md p-2 text-sm bg-gray-50"
                >
                  <div className="flex flex-wrap items-center gap-2 mb-1">
                    <span className="font-mono text-xs">{job.id}</span>
                    <span
                      className={`text-white text-xs px-2 py-0.5 rounded ${scenarioGenerationStatusBadgeClass(job.status)}`}
                    >
                      {job.status}
                    </span>
                    <span>{job.openEnded ? 'Open-Ended' : 'Choice'}</span>
                  </div>
                  <div className="text-gray-700">Zone: {zoneName}</div>
                  <div className="text-gray-700">
                    Location:{' '}
                    {typeof job.latitude === 'number' &&
                    typeof job.longitude === 'number'
                      ? `${job.latitude.toFixed(5)}, ${job.longitude.toFixed(5)}`
                      : 'Auto-selected'}
                  </div>
                  <div className="text-gray-600 text-xs">
                    Created: {formatDate(job.createdAt)}
                  </div>
                  {job.generatedScenarioId && (
                    <div className="text-gray-700">
                      Scenario ID:{' '}
                      <span className="font-mono text-xs">
                        {job.generatedScenarioId}
                      </span>
                    </div>
                  )}
                  {job.errorMessage && (
                    <div className="text-red-600 text-xs mt-1">
                      {job.errorMessage}
                    </div>
                  )}
                  {record && (
                    <div className="mt-2">
                      <button
                        type="button"
                        className="bg-blue-500 text-white px-2 py-1 rounded-md text-xs"
                        onClick={() => openEdit(record)}
                      >
                        Open Scenario
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {error && <div className="mb-3 text-red-600">{error}</div>}

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search scenarios by prompt, zone, or ID..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      <div
        className="grid gap-4"
        style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))' }}
      >
        {filtered.map((record) => {
          const zoneName =
            zones.find((zone) => zone.id === record.zoneId)?.name ??
            record.zoneId;
          return (
            <div
              key={record.id}
              className="border rounded-md p-4 bg-white shadow-sm"
            >
              <div className="text-xs text-gray-500 mb-1">{record.id}</div>
              <div className="font-semibold mb-2">
                {record.openEnded ? 'Open-Ended' : 'Choice'} Scenario
              </div>
              <div className="text-sm text-gray-700 mb-1">Zone: {zoneName}</div>
              <div className="text-sm text-gray-700 mb-1">
                Location: {record.latitude.toFixed(5)},{' '}
                {record.longitude.toFixed(5)}
              </div>
              <div className="text-sm text-gray-700 mb-2">
                Difficulty: {record.difficulty}
              </div>
              <div className="text-sm text-gray-800 mb-3 line-clamp-3">
                {record.prompt}
              </div>
              {(record.thumbnailUrl || record.imageUrl) && (
                <button
                  type="button"
                  className="w-full mb-3"
                  onClick={() =>
                    setExpandedScenarioImage({
                      url: record.imageUrl || record.thumbnailUrl,
                      title: record.prompt,
                    })
                  }
                >
                  <img
                    src={record.thumbnailUrl || record.imageUrl}
                    alt="Scenario"
                    className="w-full aspect-square object-cover rounded-md border cursor-zoom-in"
                  />
                </button>
              )}
              <div className="flex gap-2">
                <button
                  className="bg-blue-500 text-white px-3 py-1 rounded-md"
                  onClick={() => openEdit(record)}
                >
                  Edit
                </button>
                <button
                  className="bg-purple-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
                  onClick={() => handleGenerateScenarioImage(record)}
                  disabled={generatingScenarioId === record.id}
                >
                  {generatingScenarioId === record.id
                    ? 'Generating…'
                    : 'Generate Image'}
                </button>
                <button
                  className="bg-red-500 text-white px-3 py-1 rounded-md"
                  onClick={() => setDeleteId(record.id)}
                >
                  Delete
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 w-full max-w-5xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">
              {editingId ? 'Edit Scenario' : 'Create Scenario'}
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <label className="text-sm">
                Zone
                <select
                  value={form.zoneId}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, zoneId: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  <option value="">Select zone</option>
                  {zones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="text-sm">
                Difficulty
                <input
                  value={form.difficulty}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, difficulty: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                />
              </label>
              <label className="text-sm">
                Latitude
                <input
                  value={form.latitude}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, latitude: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <label className="text-sm">
                Longitude
                <input
                  value={form.longitude}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, longitude: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <div className="text-sm md:col-span-2">
                <button
                  type="button"
                  className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                  onClick={handleUseCurrentLocation}
                  disabled={geoLoading}
                >
                  {geoLoading ? 'Locating…' : 'Use Current Browser Location'}
                </button>
              </div>
              <div className="text-sm md:col-span-2">
                <div className="flex items-center justify-between mb-1">
                  <span>Map Location Picker</span>
                  <span className="text-xs text-gray-500">
                    Click map to set latitude/longitude
                  </span>
                </div>
                {mapboxgl.accessToken ? (
                  <div
                    ref={mapContainerRef}
                    className="w-full h-64 border rounded-md"
                  />
                ) : (
                  <div className="w-full border rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                    Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                    unavailable.
                  </div>
                )}
              </div>
              <label className="text-sm md:col-span-2">
                Image URL
                <input
                  value={form.imageUrl}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, imageUrl: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                />
              </label>
              <label className="text-sm md:col-span-2">
                Thumbnail URL
                <input
                  value={form.thumbnailUrl}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      thumbnailUrl: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                />
              </label>
            </div>

            <label className="text-sm block mb-4">
              Prompt
              <textarea
                value={form.prompt}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, prompt: e.target.value }))
                }
                className="w-full border rounded-md p-2 min-h-[90px]"
              />
            </label>

            <label className="inline-flex items-center gap-2 mb-4">
              <input
                type="checkbox"
                checked={form.openEnded}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    openEnded: e.target.checked,
                    failurePenaltyMode: e.target.checked
                      ? 'shared'
                      : prev.failurePenaltyMode,
                    successRewardMode: e.target.checked
                      ? 'shared'
                      : prev.successRewardMode,
                    options: e.target.checked
                      ? prev.options
                      : prev.options.length > 0
                        ? prev.options
                        : [emptyOption()],
                  }))
                }
              />
              Open-ended scenario (freeform response)
            </label>

            {form.openEnded ? (
              <div className="border rounded-md p-3 mb-4">
                <div className="font-medium mb-2">Scenario Rewards</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                  <label className="text-sm">
                    Reward Experience
                    <input
                      value={form.rewardExperience}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          rewardExperience: e.target.value,
                        }))
                      }
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                    />
                  </label>
                  <label className="text-sm">
                    Reward Gold
                    <input
                      value={form.rewardGold}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          rewardGold: e.target.value,
                        }))
                      }
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                    />
                  </label>
                </div>
                <div className="flex items-center justify-between mb-2">
                  <div className="font-medium">Item Rewards</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addScenarioItemReward}
                  >
                    Add Item
                  </button>
                </div>
                {form.itemRewards.map((reward, index) => (
                  <div
                    key={index}
                    className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2"
                  >
                    <select
                      value={reward.inventoryItemId}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          inventoryItemId: parseIntValue(e.target.value, 0),
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value={0}>Select item</option>
                      {inventoryItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                    <input
                      value={reward.quantity}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          quantity: parseIntValue(e.target.value, 1),
                        })
                      }
                      className="border rounded-md p-2"
                      type="number"
                      min={1}
                    />
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      onClick={() => removeScenarioItemReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}

                <div className="flex items-center justify-between mb-2 mt-4">
                  <div className="font-medium">Spell Rewards</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addScenarioSpellReward}
                  >
                    Add Spell
                  </button>
                </div>
                {form.spellRewards.map((reward, index) => (
                  <div
                    key={`scenario-spell-${index}`}
                    className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2"
                  >
                    <select
                      value={reward.spellId}
                      onChange={(e) =>
                        updateScenarioSpellReward(index, {
                          spellId: e.target.value,
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value="">Select spell</option>
                      {spells.map((spell) => (
                        <option key={spell.id} value={spell.id}>
                          {spell.name}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      onClick={() => removeScenarioSpellReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}

                {renderFailurePenaltyEditor({
                  title: 'Failure Penalty (Open-Ended)',
                  healthDrainType: form.failureHealthDrainType,
                  healthDrainValue: parseIntValue(
                    form.failureHealthDrainValue,
                    0
                  ),
                  manaDrainType: form.failureManaDrainType,
                  manaDrainValue: parseIntValue(form.failureManaDrainValue, 0),
                  onHealthDrainTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainType: value,
                    })),
                  onHealthDrainValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainValue: String(value),
                    })),
                  onManaDrainTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainType: value,
                    })),
                  onManaDrainValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainValue: String(value),
                    })),
                  statuses: form.failureStatuses,
                  onAddStatus: addScenarioFailureStatus,
                  onUpdateStatus: updateScenarioFailureStatus,
                  onRemoveStatus: removeScenarioFailureStatus,
                })}

                {renderSuccessRewardEditor({
                  title: 'Success Rewards (Open-Ended)',
                  healthRestoreType: form.successHealthRestoreType,
                  healthRestoreValue: parseIntValue(
                    form.successHealthRestoreValue,
                    0
                  ),
                  manaRestoreType: form.successManaRestoreType,
                  manaRestoreValue: parseIntValue(
                    form.successManaRestoreValue,
                    0
                  ),
                  onHealthRestoreTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreType: value,
                    })),
                  onHealthRestoreValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreValue: String(value),
                    })),
                  onManaRestoreTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreType: value,
                    })),
                  onManaRestoreValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreValue: String(value),
                    })),
                  statuses: form.successStatuses,
                  onAddStatus: addScenarioSuccessStatus,
                  onUpdateStatus: updateScenarioSuccessStatus,
                  onRemoveStatus: removeScenarioSuccessStatus,
                })}
              </div>
            ) : (
              <div className="border rounded-md p-3 mb-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="font-medium">Response Options</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addOption}
                  >
                    Add Option
                  </button>
                </div>

                <label className="text-sm block mb-3">
                  Failure Penalty Mode
                  <select
                    value={form.failurePenaltyMode}
                    onChange={(e) =>
                      setForm((prev) => ({
                        ...prev,
                        failurePenaltyMode: e.target
                          .value as ScenarioFailurePenaltyMode,
                      }))
                    }
                    className="w-full border rounded-md p-2"
                  >
                    <option value="shared">Shared (scenario-level)</option>
                    <option value="individual">Individual (per option)</option>
                  </select>
                </label>

                {form.failurePenaltyMode === 'shared' &&
                  renderFailurePenaltyEditor({
                    title: 'Shared Failure Penalty',
                    healthDrainType: form.failureHealthDrainType,
                    healthDrainValue: parseIntValue(
                      form.failureHealthDrainValue,
                      0
                    ),
                    manaDrainType: form.failureManaDrainType,
                    manaDrainValue: parseIntValue(
                      form.failureManaDrainValue,
                      0
                    ),
                    onHealthDrainTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureHealthDrainType: value,
                      })),
                    onHealthDrainValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureHealthDrainValue: String(value),
                      })),
                    onManaDrainTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureManaDrainType: value,
                      })),
                    onManaDrainValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureManaDrainValue: String(value),
                      })),
                    statuses: form.failureStatuses,
                    onAddStatus: addScenarioFailureStatus,
                    onUpdateStatus: updateScenarioFailureStatus,
                    onRemoveStatus: removeScenarioFailureStatus,
                  })}

                <label className="text-sm block mb-3">
                  Success Reward Mode
                  <select
                    value={form.successRewardMode}
                    onChange={(e) =>
                      setForm((prev) => ({
                        ...prev,
                        successRewardMode: e.target
                          .value as ScenarioSuccessRewardMode,
                      }))
                    }
                    className="w-full border rounded-md p-2"
                  >
                    <option value="shared">Shared (scenario-level)</option>
                    <option value="individual">Individual (per option)</option>
                  </select>
                </label>

                {form.successRewardMode === 'shared' &&
                  renderSuccessRewardEditor({
                    title: 'Shared Success Rewards',
                    healthRestoreType: form.successHealthRestoreType,
                    healthRestoreValue: parseIntValue(
                      form.successHealthRestoreValue,
                      0
                    ),
                    manaRestoreType: form.successManaRestoreType,
                    manaRestoreValue: parseIntValue(
                      form.successManaRestoreValue,
                      0
                    ),
                    onHealthRestoreTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successHealthRestoreType: value,
                      })),
                    onHealthRestoreValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successHealthRestoreValue: String(value),
                      })),
                    onManaRestoreTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successManaRestoreType: value,
                      })),
                    onManaRestoreValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successManaRestoreValue: String(value),
                      })),
                    statuses: form.successStatuses,
                    onAddStatus: addScenarioSuccessStatus,
                    onUpdateStatus: updateScenarioSuccessStatus,
                    onRemoveStatus: removeScenarioSuccessStatus,
                  })}

                {form.options.map((option, optionIndex) => (
                  <div key={optionIndex} className="border rounded-md p-3 mb-3">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <label className="text-sm md:col-span-2">
                        Option Text
                        <input
                          value={option.optionText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              optionText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Success Text
                        <textarea
                          value={option.successText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              successText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Failure Text
                        <textarea
                          value={option.failureText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              failureText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                      </label>
                      <label className="text-sm">
                        Stat Tag
                        <select
                          value={option.statTag}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              statTag: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2"
                        >
                          {statTags.map((tag) => (
                            <option key={tag} value={tag}>
                              {tag}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="text-sm">
                        Difficulty Override (optional)
                        <input
                          value={option.difficulty ?? ''}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              difficulty:
                                e.target.value === ''
                                  ? null
                                  : parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Proficiencies (comma separated)
                        <input
                          value={option.proficiencies.join(', ')}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              proficiencies: parseCsv(e.target.value),
                            })
                          }
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm">
                        Reward Experience
                        <input
                          value={option.rewardExperience}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardExperience: parseIntValue(
                                e.target.value,
                                0
                              ),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm">
                        Reward Gold
                        <input
                          value={option.rewardGold}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardGold: parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                    </div>

                    <div className="mt-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="font-medium text-sm">
                          Option Item Rewards
                        </div>
                        <button
                          type="button"
                          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                          onClick={() => addOptionItemReward(optionIndex)}
                        >
                          Add Item
                        </button>
                      </div>
                      {option.itemRewards.map((reward, rewardIndex) => (
                        <div
                          key={rewardIndex}
                          className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2"
                        >
                          <select
                            value={reward.inventoryItemId}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                inventoryItemId: parseIntValue(
                                  e.target.value,
                                  0
                                ),
                              })
                            }
                            className="border rounded-md p-2"
                          >
                            <option value={0}>Select item</option>
                            {inventoryItems.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name}
                              </option>
                            ))}
                          </select>
                          <input
                            value={reward.quantity}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                quantity: parseIntValue(e.target.value, 1),
                              })
                            }
                            className="border rounded-md p-2"
                            type="number"
                            min={1}
                          />
                          <button
                            type="button"
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                            onClick={() =>
                              removeOptionItemReward(optionIndex, rewardIndex)
                            }
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>

                    <div className="mt-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="font-medium text-sm">
                          Option Spell Rewards
                        </div>
                        <button
                          type="button"
                          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                          onClick={() => addOptionSpellReward(optionIndex)}
                        >
                          Add Spell
                        </button>
                      </div>
                      {option.spellRewards.map((reward, rewardIndex) => (
                        <div
                          key={rewardIndex}
                          className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2"
                        >
                          <select
                            value={reward.spellId}
                            onChange={(e) =>
                              updateOptionSpellReward(
                                optionIndex,
                                rewardIndex,
                                {
                                  spellId: e.target.value,
                                }
                              )
                            }
                            className="border rounded-md p-2"
                          >
                            <option value="">Select spell</option>
                            {spells.map((spell) => (
                              <option key={spell.id} value={spell.id}>
                                {spell.name}
                              </option>
                            ))}
                          </select>
                          <button
                            type="button"
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                            onClick={() =>
                              removeOptionSpellReward(optionIndex, rewardIndex)
                            }
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>

                    {form.failurePenaltyMode === 'individual' &&
                      renderFailurePenaltyEditor({
                        title: 'Option Failure Penalty',
                        healthDrainType: option.failureHealthDrainType,
                        healthDrainValue: option.failureHealthDrainValue,
                        manaDrainType: option.failureManaDrainType,
                        manaDrainValue: option.failureManaDrainValue,
                        onHealthDrainTypeChange: (value) =>
                          updateOption(optionIndex, {
                            failureHealthDrainType: value,
                          }),
                        onHealthDrainValueChange: (value) =>
                          updateOption(optionIndex, {
                            failureHealthDrainValue: value,
                          }),
                        onManaDrainTypeChange: (value) =>
                          updateOption(optionIndex, {
                            failureManaDrainType: value,
                          }),
                        onManaDrainValueChange: (value) =>
                          updateOption(optionIndex, {
                            failureManaDrainValue: value,
                          }),
                        statuses: option.failureStatuses,
                        onAddStatus: () => addOptionFailureStatus(optionIndex),
                        onUpdateStatus: (statusIndex, next) =>
                          updateOptionFailureStatus(
                            optionIndex,
                            statusIndex,
                            next
                          ),
                        onRemoveStatus: (statusIndex) =>
                          removeOptionFailureStatus(optionIndex, statusIndex),
                      })}

                    {form.successRewardMode === 'individual' &&
                      renderSuccessRewardEditor({
                        title: 'Option Success Rewards',
                        healthRestoreType: option.successHealthRestoreType,
                        healthRestoreValue: option.successHealthRestoreValue,
                        manaRestoreType: option.successManaRestoreType,
                        manaRestoreValue: option.successManaRestoreValue,
                        onHealthRestoreTypeChange: (value) =>
                          updateOption(optionIndex, {
                            successHealthRestoreType: value,
                          }),
                        onHealthRestoreValueChange: (value) =>
                          updateOption(optionIndex, {
                            successHealthRestoreValue: value,
                          }),
                        onManaRestoreTypeChange: (value) =>
                          updateOption(optionIndex, {
                            successManaRestoreType: value,
                          }),
                        onManaRestoreValueChange: (value) =>
                          updateOption(optionIndex, {
                            successManaRestoreValue: value,
                          }),
                        statuses: option.successStatuses,
                        onAddStatus: () => addOptionSuccessStatus(optionIndex),
                        onUpdateStatus: (statusIndex, next) =>
                          updateOptionSuccessStatus(
                            optionIndex,
                            statusIndex,
                            next
                          ),
                        onRemoveStatus: (statusIndex) =>
                          removeOptionSuccessStatus(optionIndex, statusIndex),
                      })}

                    <div className="mt-3">
                      <button
                        type="button"
                        className="bg-red-500 text-white px-3 py-1 rounded-md"
                        onClick={() => removeOption(optionIndex)}
                      >
                        Remove Option
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              <button
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
                onClick={save}
              >
                {editingId ? 'Update Scenario' : 'Create Scenario'}
              </button>
              <button
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                onClick={closeModal}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {deleteId && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 max-w-sm w-full">
            <h3 className="text-lg font-semibold mb-2">Delete Scenario</h3>
            <p className="text-sm text-gray-700 mb-4">
              This action cannot be undone.
            </p>
            <div className="flex gap-2">
              <button
                className="bg-red-500 text-white px-4 py-2 rounded-md"
                onClick={confirmDelete}
              >
                Delete
              </button>
              <button
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                onClick={() => setDeleteId(null)}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {expandedScenarioImage && (
        <div
          className="fixed inset-0 bg-black/80 z-[60] flex items-center justify-center p-4"
          onClick={() => setExpandedScenarioImage(null)}
        >
          <div
            className="max-w-6xl max-h-[90vh] w-full"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex justify-end mb-2">
              <button
                type="button"
                className="bg-white text-gray-900 px-3 py-1 rounded-md"
                onClick={() => setExpandedScenarioImage(null)}
              >
                Close
              </button>
            </div>
            <img
              src={expandedScenarioImage.url}
              alt={expandedScenarioImage.title}
              className="w-full max-h-[80vh] object-contain rounded-md bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
};
