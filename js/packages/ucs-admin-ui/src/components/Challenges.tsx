import React, {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { PointOfInterest } from '@poltergeist/types';
import {
  MaterialRewardsEditor,
  MaterialRewardForm,
  summarizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';
import { useSearchParams } from 'react-router-dom';

type ChallengeRecord = {
  id: string;
  zoneId: string;
  pointOfInterestId?: string | null;
  latitude: number;
  longitude: number;
  polygonPoints?: [number, number][] | null;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  reward: number;
  materialRewards?: MaterialRewardForm[];
  inventoryItemId?: number | null;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: number;
  scaleWithUserLevel: boolean;
  recurrenceFrequency?: string | null;
  nextRecurrenceAt?: string | null;
  statTags: string[];
  proficiency?: string | null;
};

type ChallengeFormState = {
  zoneId: string;
  locationMode: 'poi' | 'coordinates' | 'polygon';
  pointOfInterestId: string;
  latitude: string;
  longitude: string;
  polygonPoints: string;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  reward: string;
  materialRewards: MaterialRewardForm[];
  inventoryItemId: string;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: string;
  scaleWithUserLevel: boolean;
  recurrenceFrequency: string;
  statTags: string;
  proficiency: string;
};

type ImagePreviewState = {
  url: string;
  alt: string;
};

type ChallengeGenerationJob = {
  id: string;
  zoneId: string;
  pointOfInterestId?: string | null;
  status: string;
  count: number;
  createdCount: number;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ChallengeGenerationFormState = {
  zoneId: string;
  pointOfInterestId: string;
  count: string;
};

type PointOfInterestOption = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
};

type SelectOption = {
  value: string;
  label: string;
  secondary?: string;
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

const challengeListPageSize = 25;

const PaginationControls = ({
  page,
  pageSize,
  total,
  label,
  onPageChange,
}: {
  page: number;
  pageSize: number;
  total: number;
  label: string;
  onPageChange: (page: number) => void;
}) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(total, page * pageSize);

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
      <p className="text-sm text-gray-600">
        {total === 0
          ? `No ${label}.`
          : `Showing ${start}-${end} of ${total} ${label}`}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
        >
          Previous
        </button>
        <span className="text-sm text-gray-600">
          Page {page} of {totalPages}
        </span>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
        >
          Next
        </button>
      </div>
    </div>
  );
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
  disabled = false,
  noMatchesLabel = 'No matches found',
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  noMatchesLabel?: string;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((option) => option.value === value);
  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return options;
    return options.filter((option) => {
      const haystack =
        `${option.label} ${option.secondary ?? ''}`.toLowerCase();
      return haystack.includes(normalized);
    });
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? '';

  return (
    <div className="relative">
      <label className="block text-sm">{label}</label>
      <input
        value={displayValue}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          if (disabled) return;
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          window.setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        disabled={disabled}
        className="w-full border rounded-md p-2 disabled:bg-gray-100 disabled:text-gray-500"
      />
      {open && !disabled ? (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 ? (
            <div className="px-3 py-2 text-sm text-gray-500">
              {noMatchesLabel}
            </div>
          ) : (
            filtered.map((option) => (
              <button
                type="button"
                key={option.value || '__empty-option__'}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(option.value);
                  setOpen(false);
                  setQuery('');
                }}
                className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-indigo-50"
              >
                <span className="font-medium text-gray-900">
                  {option.label}
                </span>
                {option.secondary ? (
                  <span className="text-xs text-gray-500">
                    {option.secondary}
                  </span>
                ) : null}
              </button>
            ))
          )}
        </div>
      ) : null}
    </div>
  );
};

const statTagOptions = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
] as const;

const recurrenceOptions = [
  { value: '', label: 'No Recurrence' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseOptionalInt = (value: string): number | undefined => {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseCsv = (value: string): string[] =>
  value
    .split(',')
    .map((entry) => entry.trim().toLowerCase())
    .filter((entry) =>
      statTagOptions.includes(entry as (typeof statTagOptions)[number])
    );

const formatDate = (value?: string | null): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const parsePolygonPoints = (
  input: string,
  minimumPoints = 3
): [number, number][] | null => {
  if (!input.trim()) return null;
  try {
    const parsed = JSON.parse(input);
    if (!Array.isArray(parsed)) return null;
    const points: [number, number][] = [];
    for (const entry of parsed) {
      if (!Array.isArray(entry) || entry.length < 2) return null;
      const lng = Number(entry[0]);
      const lat = Number(entry[1]);
      if (!Number.isFinite(lng) || !Number.isFinite(lat)) return null;
      points.push([lng, lat]);
    }
    return points.length >= minimumPoints ? points : null;
  } catch {
    return null;
  }
};

const formatPolygonPoints = (points?: [number, number][] | null): string => {
  if (!points || points.length === 0) return '';
  return JSON.stringify(points);
};

const polygonCenter = (
  points?: [number, number][] | null
): [number, number] | null => {
  if (!points || points.length === 0) return null;
  let lngSum = 0;
  let latSum = 0;
  points.forEach(([lng, lat]) => {
    lngSum += lng;
    latSum += lat;
  });
  return [lngSum / points.length, latSum / points.length];
};

const emptyForm = (): ChallengeFormState => ({
  zoneId: '',
  locationMode: 'coordinates',
  pointOfInterestId: '',
  latitude: '',
  longitude: '',
  polygonPoints: '',
  question: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  reward: '0',
  materialRewards: [],
  inventoryItemId: '',
  submissionType: 'photo',
  difficulty: '0',
  scaleWithUserLevel: false,
  recurrenceFrequency: '',
  statTags: '',
  proficiency: '',
});

const emptyGenerationForm = (): ChallengeGenerationFormState => ({
  zoneId: '',
  pointOfInterestId: '',
  count: '6',
});

const challengeGenerationStatusBadgeClass = (status: string): string => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    case 'in_progress':
      return 'bg-amber-600';
    default:
      return 'bg-gray-600';
  }
};

const formFromRecord = (record: ChallengeRecord): ChallengeFormState => ({
  zoneId: record.zoneId ?? '',
  locationMode:
    record.polygonPoints && record.polygonPoints.length >= 3
      ? 'polygon'
      : record.pointOfInterestId
        ? 'poi'
        : 'coordinates',
  pointOfInterestId: record.pointOfInterestId ?? '',
  latitude: String(record.latitude ?? ''),
  longitude: String(record.longitude ?? ''),
  polygonPoints: formatPolygonPoints(record.polygonPoints),
  question: record.question ?? '',
  description: record.description ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  rewardExperience: String(record.rewardExperience ?? 0),
  reward: String(record.reward ?? 0),
  materialRewards: (record.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount ?? 1,
  })),
  inventoryItemId:
    record.inventoryItemId !== undefined && record.inventoryItemId !== null
      ? String(record.inventoryItemId)
      : '',
  submissionType:
    record.submissionType === 'text' || record.submissionType === 'video'
      ? record.submissionType
      : 'photo',
  difficulty: String(record.difficulty ?? 0),
  scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
  recurrenceFrequency: record.recurrenceFrequency ?? '',
  statTags: (record.statTags ?? []).join(', '),
  proficiency: record.proficiency ?? '',
});

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Challenges = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();
  const [zonePointOfInterestMap, setZonePointOfInterestMap] = useState<
    Record<string, PointOfInterestOption[]>
  >({});
  const [pointOfInterestLoadingByZone, setPointOfInterestLoadingByZone] =
    useState<Record<string, boolean>>({});

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [records, setRecords] = useState<ChallengeRecord[]>([]);
  const [query, setQuery] = useState('');
  const [zoneQuery, setZoneQuery] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showModal, setShowModal] = useState(false);
  const [editingChallenge, setEditingChallenge] =
    useState<ChallengeRecord | null>(null);
  const [form, setForm] = useState<ChallengeFormState>(emptyForm);
  const [geoLoading, setGeoLoading] = useState(false);
  const [generatingChallengeId, setGeneratingChallengeId] = useState<
    string | null
  >(null);
  const [imagePreview, setImagePreview] = useState<ImagePreviewState | null>(
    null
  );
  const [generationForm, setGenerationForm] =
    useState<ChallengeGenerationFormState>(emptyGenerationForm);
  const [generationJobs, setGenerationJobs] = useState<
    ChallengeGenerationJob[]
  >([]);
  const [generationJobsLoading, setGenerationJobsLoading] = useState(false);
  const [generationSubmitting, setGenerationSubmitting] = useState(false);
  const [generationError, setGenerationError] = useState<string | null>(null);
  const [bulkDeletingChallenges, setBulkDeletingChallenges] = useState(false);
  const [selectedChallengeIds, setSelectedChallengeIds] = useState<Set<string>>(
    new Set()
  );
  const seenCompletedGenerationJobsRef = React.useRef<Set<string>>(new Set());
  const didHydrateDeepLinkedChallengeRef = React.useRef(false);
  const deepLinkedChallengeId = searchParams.get('id')?.trim() ?? '';
  const replaceDeepLinkedChallengeId = useCallback((challengeId?: string | null) => {
    const normalizedChallengeId = (challengeId ?? '').trim();
    const currentChallengeId = searchParams.get('id')?.trim() ?? '';
    if (normalizedChallengeId === currentChallengeId) {
      return;
    }
    const next = new URLSearchParams(searchParams);
    if (normalizedChallengeId) {
      next.set('id', normalizedChallengeId);
    } else {
      next.delete('id');
    }
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(form.latitude);
  const formLongitudeRef = React.useRef(form.longitude);
  const formLocationModeRef = React.useRef(form.locationMode);
  const [polygonDraftPoints, setPolygonDraftPoints] = useState<
    [number, number][]
  >([]);
  const deferredQuery = useDeferredValue(query);
  const deferredZoneQuery = useDeferredValue(zoneQuery);

  const loadPointsOfInterestForZone = useCallback(
    async (zoneId: string) => {
      const trimmedZoneId = zoneId.trim();
      if (!trimmedZoneId) return;
      if (zonePointOfInterestMap[trimmedZoneId]) return;
      setPointOfInterestLoadingByZone((prev) => ({
        ...prev,
        [trimmedZoneId]: true,
      }));
      try {
        const points = await apiClient.get<PointOfInterest[]>(
          `/sonar/zones/${trimmedZoneId}/pointsOfInterest`
        );
        const mapped = (Array.isArray(points) ? points : [])
          .map((point) => {
            const lat = Number.parseFloat(String(point.lat ?? ''));
            const lng = Number.parseFloat(String(point.lng ?? ''));
            if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
            return {
              id: point.id,
              name: point.name || point.id,
              latitude: lat,
              longitude: lng,
            };
          })
          .filter((point): point is PointOfInterestOption => point !== null);
        setZonePointOfInterestMap((prev) => ({
          ...prev,
          [trimmedZoneId]: mapped,
        }));
      } catch (err) {
        console.error(
          `Failed to load points of interest for zone ${trimmedZoneId}`,
          err
        );
      } finally {
        setPointOfInterestLoadingByZone((prev) => ({
          ...prev,
          [trimmedZoneId]: false,
        }));
      }
    },
    [apiClient, zonePointOfInterestMap]
  );

  const zoneNameById = useMemo(() => {
    return new Map(zones.map((zone) => [zone.id, zone.name]));
  }, [zones]);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<PaginatedResponse<ChallengeRecord>>(
        '/sonar/admin/challenges',
        {
          page,
          pageSize: challengeListPageSize,
          query: deferredQuery.trim(),
          zoneQuery: deferredZoneQuery.trim(),
        }
      );
      setRecords(Array.isArray(response?.items) ? response.items : []);
      setTotal(response?.total ?? 0);
    } catch (err) {
      console.error('Failed to load challenges', err);
      setError('Failed to load challenges.');
    } finally {
      setLoading(false);
    }
  }, [apiClient, deferredQuery, deferredZoneQuery, page]);

  const refreshChallengeById = useCallback(
    async (challengeId: string) => {
      const latest = await apiClient.get<ChallengeRecord>(
        `/sonar/challenges/${challengeId}`
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === challengeId ? latest : record))
      );
      return latest;
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    setPage(1);
  }, [query, zoneQuery]);

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(total / challengeListPageSize));
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [page, total]);

  const loadGenerationJobs = useCallback(async () => {
    try {
      setGenerationJobsLoading(true);
      const response = await apiClient.get<ChallengeGenerationJob[]>(
        '/sonar/admin/challenge-generation-jobs',
        { limit: 25 }
      );
      setGenerationJobs(Array.isArray(response) ? response : []);
      setGenerationError(null);
    } catch (err) {
      console.error('Failed to load challenge generation jobs', err);
      setGenerationError('Failed to load challenge generation jobs.');
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
    const completed = generationJobs.filter(
      (job) => job.status === 'completed'
    );
    let shouldReloadChallenges = false;
    for (const job of completed) {
      if (!seenCompletedGenerationJobsRef.current.has(job.id)) {
        seenCompletedGenerationJobsRef.current.add(job.id);
        if (job.createdCount > 0) {
          shouldReloadChallenges = true;
        }
      }
    }
    if (shouldReloadChallenges) {
      void load();
    }
  }, [generationJobs, load]);
  const filteredRecords = records;
  const selectedChallengeIdSet = useMemo(
    () => selectedChallengeIds,
    [selectedChallengeIds]
  );
  const allFilteredChallengesSelected = useMemo(() => {
    if (filteredRecords.length === 0) return false;
    return filteredRecords.every((record) =>
      selectedChallengeIds.has(record.id)
    );
  }, [filteredRecords, selectedChallengeIds]);

  useEffect(() => {
    setSelectedChallengeIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(records.map((record) => record.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [records]);

  const allPointOfInterestNamesById = useMemo(() => {
    const byId = new Map<string, string>();
    Object.values(zonePointOfInterestMap).forEach((points) => {
      points.forEach((point) => {
        if (!byId.has(point.id)) {
          byId.set(point.id, point.name);
        }
      });
    });
    return byId;
  }, [zonePointOfInterestMap]);

  const pointsOfInterestForFormZone = useMemo(() => {
    return zonePointOfInterestMap[form.zoneId] ?? [];
  }, [form.zoneId, zonePointOfInterestMap]);
  const pointsOfInterestForGenerationZone = useMemo(() => {
    return zonePointOfInterestMap[generationForm.zoneId] ?? [];
  }, [generationForm.zoneId, zonePointOfInterestMap]);
  const generationZoneOptions = useMemo<SelectOption[]>(
    () => [
      { value: '', label: 'Select zone' },
      ...zones.map((zone) => ({
        value: zone.id,
        label: zone.name || zone.id,
        secondary: zone.id,
      })),
    ],
    [zones]
  );
  const generationPointOfInterestOptions = useMemo<SelectOption[]>(
    () => [
      { value: '', label: 'Any POI in zone' },
      ...pointsOfInterestForGenerationZone.map((point) => ({
        value: point.id,
        label: point.name || point.id,
        secondary: point.id,
      })),
    ],
    [pointsOfInterestForGenerationZone]
  );
  const hasSelectedPointOfInterest =
    form.locationMode === 'poi' && form.pointOfInterestId.trim().length > 0;
  const isCoordinateMode = form.locationMode === 'coordinates';
  const isPolygonMode = form.locationMode === 'polygon';
  const usesInteractiveMap = isCoordinateMode || isPolygonMode;

  useEffect(() => {
    if (!showModal) return;
    if (!form.zoneId) return;
    void loadPointsOfInterestForZone(form.zoneId);
  }, [form.zoneId, loadPointsOfInterestForZone, showModal]);

  useEffect(() => {
    if (form.locationMode !== 'polygon') {
      setPolygonDraftPoints([]);
      return;
    }
    setPolygonDraftPoints(parsePolygonPoints(form.polygonPoints, 1) ?? []);
  }, [form.locationMode, form.polygonPoints]);

  useEffect(() => {
    if (!generationForm.zoneId) return;
    void loadPointsOfInterestForZone(generationForm.zoneId);
  }, [generationForm.zoneId, loadPointsOfInterestForZone]);

  useEffect(() => {
    if (!records.length) return;
    const zoneIds = Array.from(new Set(records.map((record) => record.zoneId)));
    zoneIds.forEach((zoneId) => {
      if (zoneId && !zonePointOfInterestMap[zoneId]) {
        void loadPointsOfInterestForZone(zoneId);
      }
    });
  }, [loadPointsOfInterestForZone, records, zonePointOfInterestMap]);

  useEffect(() => {
    if (!generationJobs.length) return;
    const zoneIds = Array.from(
      new Set(generationJobs.map((job) => job.zoneId))
    );
    zoneIds.forEach((zoneId) => {
      if (zoneId && !zonePointOfInterestMap[zoneId]) {
        void loadPointsOfInterestForZone(zoneId);
      }
    });
  }, [generationJobs, loadPointsOfInterestForZone, zonePointOfInterestMap]);

  const openCreate = () => {
    setEditingChallenge(null);
    setForm({
      ...emptyForm(),
      zoneId: zones[0]?.id ?? '',
    });
    setShowModal(true);
  };

  const openEdit = (record: ChallengeRecord) => {
    setEditingChallenge(record);
    setForm(formFromRecord(record));
    setShowModal(true);
  };

  useEffect(() => {
    if (didHydrateDeepLinkedChallengeRef.current) {
      return;
    }
    if (!deepLinkedChallengeId) {
      didHydrateDeepLinkedChallengeRef.current = true;
      return;
    }
    if (editingChallenge?.id === deepLinkedChallengeId && showModal) {
      didHydrateDeepLinkedChallengeRef.current = true;
      return;
    }
    void apiClient
      .get<ChallengeRecord>(`/sonar/challenges/${deepLinkedChallengeId}`)
      .then((record) => {
        if (!record) {
          didHydrateDeepLinkedChallengeRef.current = true;
          return;
        }
        setRecords((prev) => {
          const withoutExisting = prev.filter((entry) => entry.id !== record.id);
          return [record, ...withoutExisting];
        });
        openEdit(record);
      })
      .catch((error) => {
        console.error('Failed to deep link challenge', error);
        didHydrateDeepLinkedChallengeRef.current = true;
      });
  }, [apiClient, deepLinkedChallengeId, editingChallenge, showModal]);

  useEffect(() => {
    if (!didHydrateDeepLinkedChallengeRef.current) {
      return;
    }
    replaceDeepLinkedChallengeId(showModal ? editingChallenge?.id ?? null : null);
  }, [editingChallenge, replaceDeepLinkedChallengeId, showModal]);

  const closeModal = () => {
    setShowModal(false);
    setEditingChallenge(null);
    setForm(emptyForm());
  };

  useEffect(() => {
    formLatitudeRef.current = form.latitude;
    formLongitudeRef.current = form.longitude;
    formLocationModeRef.current = form.locationMode;
  }, [form.latitude, form.locationMode, form.longitude]);

  const setFormLocation = useCallback((latitude: number, longitude: number) => {
    setForm((prev) => ({
      ...prev,
      locationMode: 'coordinates',
      pointOfInterestId: '',
      latitude: latitude.toFixed(6),
      longitude: longitude.toFixed(6),
    }));
  }, []);

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
      (geoError) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${geoError.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setFormLocation]);

  const ensurePolygonDraftLayers = useCallback((map: mapboxgl.Map) => {
    if (!map.isStyleLoaded()) return;

    if (!map.getSource('challenge-draft-line')) {
      map.addSource('challenge-draft-line', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'LineString', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'challenge-draft-line',
        type: 'line',
        source: 'challenge-draft-line',
        paint: {
          'line-color': '#0f766e',
          'line-width': 2,
          'line-dasharray': [2, 2],
        },
      });
    }

    if (!map.getSource('challenge-draft-polygon')) {
      map.addSource('challenge-draft-polygon', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'challenge-draft-polygon',
        type: 'fill',
        source: 'challenge-draft-polygon',
        paint: {
          'fill-color': '#2563eb',
          'fill-opacity': 0.16,
        },
      });
      map.addLayer({
        id: 'challenge-draft-polygon-outline',
        type: 'line',
        source: 'challenge-draft-polygon',
        paint: {
          'line-color': '#1d4ed8',
          'line-width': 2,
        },
      });
    }
  }, []);

  useEffect(() => {
    if (!showModal) return;
    if (!usesInteractiveMap || hasSelectedPointOfInterest) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const polygonCenterPoint =
      formLocationModeRef.current === 'polygon'
        ? polygonCenter(parsePolygonPoints(form.polygonPoints, 1))
        : null;
    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === form.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] = polygonCenterPoint
      ? polygonCenterPoint
      : Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
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

    const initializePolygonLayers = () => {
      ensurePolygonDraftLayers(map);
      const draftPoints =
        formLocationModeRef.current === 'polygon'
          ? parsePolygonPoints(form.polygonPoints, 1) ?? []
          : [];
      const lineSource = map.getSource(
        'challenge-draft-line'
      ) as mapboxgl.GeoJSONSource | null;
      const polygonSource = map.getSource(
        'challenge-draft-polygon'
      ) as mapboxgl.GeoJSONSource | null;
      if (lineSource) {
        lineSource.setData({
          type: 'Feature',
          geometry: { type: 'LineString', coordinates: draftPoints },
          properties: {},
        });
      }
      if (polygonSource) {
        polygonSource.setData({
          type: 'Feature',
          geometry: {
            type: 'Polygon',
            coordinates:
              draftPoints.length >= 3 ? [[...draftPoints, draftPoints[0]]] : [],
          },
          properties: {},
        });
      }
      if (draftPoints.length > 0) {
        const bounds = new mapboxgl.LngLatBounds();
        draftPoints.forEach(([lng, lat]) => bounds.extend([lng, lat]));
        map.fitBounds(bounds, { padding: 48, duration: 0, maxZoom: 15 });
      }
    };
    if (map.isStyleLoaded()) {
      initializePolygonLayers();
    } else {
      map.on('load', initializePolygonLayers);
    }

    map.on('click', (event) => {
      if (formLocationModeRef.current === 'polygon') {
        setPolygonDraftPoints((prev) => {
          const next = [...prev, [event.lngLat.lng, event.lngLat.lat]] as [
            number,
            number,
          ][];
          setForm((formPrev) => ({
            ...formPrev,
            pointOfInterestId: '',
            polygonPoints: JSON.stringify(next),
          }));
          return next;
        });
        return;
      }
      setFormLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.off('load', initializePolygonLayers);
      map.remove();
      mapRef.current = null;
    };
  }, [
    form.zoneId,
    ensurePolygonDraftLayers,
    form.polygonPoints,
    hasSelectedPointOfInterest,
    setFormLocation,
    showModal,
    usesInteractiveMap,
    zones,
  ]);

  useEffect(() => {
    if (!showModal) return;
    if (!usesInteractiveMap) return;
    if (!mapRef.current) return;
    window.setTimeout(() => mapRef.current?.resize(), 0);
  }, [isCoordinateMode, isPolygonMode, showModal, usesInteractiveMap]);

  useEffect(() => {
    if (!showModal) return;
    if (!mapRef.current) return;

    const map = mapRef.current;
    const lineSource = map.getSource(
      'challenge-draft-line'
    ) as mapboxgl.GeoJSONSource | null;
    const polygonSource = map.getSource(
      'challenge-draft-polygon'
    ) as mapboxgl.GeoJSONSource | null;

    if (lineSource) {
      lineSource.setData({
        type: 'Feature',
        geometry: {
          type: 'LineString',
          coordinates: isPolygonMode ? polygonDraftPoints : [],
        },
        properties: {},
      });
    }

    if (polygonSource) {
      polygonSource.setData({
        type: 'Feature',
        geometry: {
          type: 'Polygon',
          coordinates:
            isPolygonMode && polygonDraftPoints.length >= 3
              ? [[...polygonDraftPoints, polygonDraftPoints[0]]]
              : [],
        },
        properties: {},
      });
    }

    if (!isCoordinateMode || hasSelectedPointOfInterest) {
      markerRef.current?.remove();
      markerRef.current = null;
      return;
    }

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

    map.easeTo({ center: [lng, lat], duration: 350 });
  }, [
    form.latitude,
    form.longitude,
    hasSelectedPointOfInterest,
    isCoordinateMode,
    isPolygonMode,
    polygonDraftPoints,
    showModal,
  ]);

  useEffect(() => {
    if (!showModal) return;
    if (!isPolygonMode) return;
    if (!mapRef.current) return;
    if (polygonDraftPoints.length === 0) return;

    const map = mapRef.current;
    const bounds = new mapboxgl.LngLatBounds();
    polygonDraftPoints.forEach(([lng, lat]) => {
      bounds.extend([lng, lat]);
    });
    map.fitBounds(bounds, { padding: 48, duration: 0, maxZoom: 15 });
  }, [isPolygonMode, polygonDraftPoints, showModal]);

  const handleQueueChallengeGeneration = async () => {
    if (!generationForm.zoneId) {
      setGenerationError('Please select a zone.');
      return;
    }
    const count = parseIntSafe(generationForm.count, 0);
    if (count < 1 || count > 100) {
      setGenerationError('Count must be between 1 and 100.');
      return;
    }

    try {
      setGenerationSubmitting(true);
      setGenerationError(null);
      const created = await apiClient.post<ChallengeGenerationJob>(
        '/sonar/admin/challenge-generation-jobs',
        {
          zoneId: generationForm.zoneId,
          pointOfInterestId: generationForm.pointOfInterestId.trim(),
          count,
        }
      );
      setGenerationJobs((prev) => [created, ...prev]);
    } catch (err) {
      console.error('Failed to queue challenge generation job', err);
      setGenerationError('Failed to queue challenge generation job.');
    } finally {
      setGenerationSubmitting(false);
    }
  };

  const save = async () => {
    try {
      const rewardMode = form.rewardMode;
      const parsedPolygonPoints =
        form.locationMode === 'polygon'
          ? parsePolygonPoints(form.polygonPoints)
          : null;
      const payload = {
        zoneId: form.zoneId.trim(),
        pointOfInterestId:
          form.locationMode === 'poi' ? form.pointOfInterestId.trim() : '',
        latitude:
          form.locationMode === 'coordinates'
            ? parseFloatSafe(form.latitude, 0)
            : 0,
        longitude:
          form.locationMode === 'coordinates'
            ? parseFloatSafe(form.longitude, 0)
            : 0,
        polygonPoints:
          form.locationMode === 'polygon' ? parsedPolygonPoints : undefined,
        question: form.question.trim(),
        description: form.description.trim(),
        imageUrl: form.imageUrl.trim(),
        thumbnailUrl: form.thumbnailUrl.trim(),
        rewardMode,
        randomRewardSize: form.randomRewardSize,
        rewardExperience:
          rewardMode === 'explicit'
            ? parseIntSafe(form.rewardExperience, 0)
            : 0,
        reward: rewardMode === 'explicit' ? parseIntSafe(form.reward, 0) : 0,
        materialRewards: rewardMode === 'explicit' ? form.materialRewards : [],
        inventoryItemId:
          rewardMode === 'explicit'
            ? parseOptionalInt(form.inventoryItemId)
            : undefined,
        submissionType: form.submissionType,
        difficulty: parseIntSafe(form.difficulty, 0),
        scaleWithUserLevel: form.scaleWithUserLevel,
        recurrenceFrequency: form.recurrenceFrequency,
        statTags: parseCsv(form.statTags),
        proficiency: form.proficiency.trim(),
      };

      if (!payload.zoneId || !payload.question) {
        alert('Zone and question are required.');
        return;
      }
      if (form.locationMode === 'poi' && !payload.pointOfInterestId) {
        alert('Select a point of interest.');
        return;
      }
      if (form.locationMode === 'polygon' && !parsedPolygonPoints) {
        alert(
          'Enter polygon points as JSON: [[lng,lat],[lng,lat],[lng,lat],...]'
        );
        return;
      }
      if (form.locationMode === 'coordinates') {
        const latitude = Number.parseFloat(form.latitude);
        const longitude = Number.parseFloat(form.longitude);
        if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
          alert(
            'Choose a point of interest or provide valid latitude/longitude.'
          );
          return;
        }
      }
      if (!payload.thumbnailUrl && payload.imageUrl) {
        payload.thumbnailUrl = payload.imageUrl;
      }

      if (editingChallenge) {
        await apiClient.put<ChallengeRecord>(
          `/sonar/challenges/${editingChallenge.id}`,
          payload
        );
      } else {
        await apiClient.post<ChallengeRecord>('/sonar/challenges', payload);
      }
      await load();
      closeModal();
    } catch (err) {
      console.error('Failed to save challenge', err);
      const message =
        err instanceof Error ? err.message : 'Failed to save challenge.';
      alert(message);
    }
  };

  const deleteChallenge = async (record: ChallengeRecord) => {
    if (bulkDeletingChallenges) return;
    if (!window.confirm(`Delete challenge "${record.question}"?`)) return;
    try {
      await apiClient.delete(`/sonar/challenges/${record.id}`);
      setSelectedChallengeIds((prev) => {
        if (!prev.has(record.id)) return prev;
        const next = new Set(prev);
        next.delete(record.id);
        return next;
      });
      await load();
    } catch (err) {
      console.error('Failed to delete challenge', err);
      alert('Failed to delete challenge.');
    }
  };

  const toggleChallengeSelection = (challengeId: string) => {
    setSelectedChallengeIds((prev) => {
      const next = new Set(prev);
      if (next.has(challengeId)) {
        next.delete(challengeId);
      } else {
        next.add(challengeId);
      }
      return next;
    });
  };

  const toggleSelectVisibleChallenges = () => {
    if (filteredRecords.length === 0) return;
    setSelectedChallengeIds((prev) => {
      const next = new Set(prev);
      if (allFilteredChallengesSelected) {
        filteredRecords.forEach((record) => next.delete(record.id));
      } else {
        filteredRecords.forEach((record) => next.add(record.id));
      }
      return next;
    });
  };

  const clearChallengeSelection = () => {
    setSelectedChallengeIds(new Set());
  };

  const handleBulkDeleteChallenges = async () => {
    if (bulkDeletingChallenges || selectedChallengeIds.size === 0) return;

    const selectedIds = Array.from(selectedChallengeIds);
    const selectedNames = records
      .filter((record) => selectedChallengeIds.has(record.id))
      .map((record) => record.question);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected challenge (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected challenges${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;
    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingChallenges(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((challengeId) =>
          apiClient.delete(`/sonar/challenges/${challengeId}`)
        )
      );
      const deletedIds = new Set<string>();
      const failedIds: string[] = [];
      results.forEach((result, index) => {
        const challengeId = selectedIds[index];
        if (result.status === 'fulfilled') {
          deletedIds.add(challengeId);
        } else {
          console.error(
            `Failed to delete challenge ${challengeId}`,
            result.reason
          );
          failedIds.push(challengeId);
        }
      });

      if (deletedIds.size > 0) {
        setSelectedChallengeIds((prev) => {
          const next = new Set(prev);
          deletedIds.forEach((challengeId) => next.delete(challengeId));
          return next;
        });
        if (editingChallenge && deletedIds.has(editingChallenge.id)) {
          closeModal();
        }
        await load();
      }

      if (failedIds.length > 0) {
        alert(
          `Deleted ${deletedIds.size} challenge${deletedIds.size === 1 ? '' : 's'}, but failed to delete ${
            failedIds.length
          }. Check console for details.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk delete challenges', err);
      alert('Failed to delete selected challenges.');
    } finally {
      setBulkDeletingChallenges(false);
    }
  };

  const handleGenerateImage = async (record: ChallengeRecord) => {
    if (generatingChallengeId) return;
    setGeneratingChallengeId(record.id);
    const previousImageURL = (
      record.imageUrl ||
      record.thumbnailUrl ||
      ''
    ).trim();
    try {
      await apiClient.post(`/sonar/challenges/${record.id}/generate-image`, {});

      for (let attempt = 0; attempt < 18; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 1200));
        const latest = await refreshChallengeById(record.id);
        const nextImageURL = (
          latest.imageUrl ||
          latest.thumbnailUrl ||
          ''
        ).trim();
        if (nextImageURL && nextImageURL !== previousImageURL) {
          break;
        }
      }
    } catch (err) {
      console.error('Failed to queue challenge image generation', err);
      alert('Failed to queue challenge image generation.');
    } finally {
      setGeneratingChallengeId(null);
    }
  };

  const openImagePreview = (record: ChallengeRecord) => {
    const url = record.thumbnailUrl || record.imageUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `Challenge ${record.question || record.id} image`,
    });
  };

  const closeImagePreview = () => {
    setImagePreview(null);
  };

  if (loading) {
    return <div className="m-10">Loading challenges...</div>;
  }

  return (
    <div className="m-10">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Challenges</h1>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={openCreate}
        >
          Create Challenge
        </button>
      </div>

      <div className="mb-4">
        <div className="flex flex-wrap gap-3">
          <input
            className="w-full max-w-xl border rounded-md p-2"
            placeholder="Search by question, description, or id"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          <input
            className="w-full max-w-sm border rounded-md p-2"
            placeholder="Search by zone"
            value={zoneQuery}
            onChange={(event) => setZoneQuery(event.target.value)}
          />
        </div>
      </div>
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
          onClick={toggleSelectVisibleChallenges}
          disabled={filteredRecords.length === 0 || bulkDeletingChallenges}
        >
          {allFilteredChallengesSelected
            ? 'Unselect Visible'
            : 'Select Visible'}
        </button>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
          onClick={clearChallengeSelection}
          disabled={selectedChallengeIds.size === 0 || bulkDeletingChallenges}
        >
          Clear Selection
        </button>
        <button
          type="button"
          className="qa-btn qa-btn-danger"
          onClick={handleBulkDeleteChallenges}
          disabled={selectedChallengeIds.size === 0 || bulkDeletingChallenges}
        >
          {bulkDeletingChallenges
            ? `Deleting ${selectedChallengeIds.size}...`
            : `Delete Selected (${selectedChallengeIds.size})`}
        </button>
      </div>

      {error ? <div className="mb-4 text-sm text-red-600">{error}</div> : null}

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">
            Generate Random Challenges (Async)
          </h2>
          <button
            type="button"
            className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
            onClick={() => void loadGenerationJobs()}
            disabled={generationJobsLoading}
          >
            {generationJobsLoading ? 'Refreshing…' : 'Refresh Jobs'}
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-3">
          <SearchableSelect
            label="Zone"
            placeholder="Search zones..."
            options={generationZoneOptions}
            value={generationForm.zoneId}
            onChange={(zoneId) =>
              setGenerationForm((prev) => ({
                ...prev,
                zoneId,
                pointOfInterestId: '',
              }))
            }
            noMatchesLabel="No zones found"
          />
          <SearchableSelect
            label="Point of Interest (Optional)"
            placeholder={
              !generationForm.zoneId
                ? 'Select a zone first'
                : pointOfInterestLoadingByZone[generationForm.zoneId]
                  ? 'Loading points of interest...'
                  : 'Search points of interest...'
            }
            options={generationPointOfInterestOptions}
            value={generationForm.pointOfInterestId}
            onChange={(pointOfInterestId) =>
              setGenerationForm((prev) => ({
                ...prev,
                pointOfInterestId,
              }))
            }
            disabled={
              !generationForm.zoneId ||
              pointOfInterestLoadingByZone[generationForm.zoneId]
            }
            noMatchesLabel="No points of interest found for this zone"
          />
          <label className="text-sm">
            Challenge Count
            <input
              value={generationForm.count}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  count: event.target.value,
                }))
              }
              className="w-full border rounded-md p-2"
              type="number"
              min={1}
              max={100}
            />
          </label>
        </div>

        <div className="flex items-center gap-2 mb-3">
          <button
            type="button"
            className="bg-indigo-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
            onClick={handleQueueChallengeGeneration}
            disabled={generationSubmitting}
          >
            {generationSubmitting ? 'Queueing…' : 'Queue Challenge Generation'}
          </button>
        </div>

        {generationError ? (
          <div className="mb-3 text-sm text-red-600">{generationError}</div>
        ) : null}

        <div className="font-medium mb-2">Recent Generation Jobs</div>
        {generationJobsLoading && generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">Loading jobs...</div>
        ) : generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">No jobs yet.</div>
        ) : (
          <div className="grid gap-2">
            {generationJobs.map((job) => (
              <div
                key={job.id}
                className="border rounded-md p-2 text-sm bg-gray-50"
              >
                <div className="flex flex-wrap items-center gap-2 mb-1">
                  <span className="font-mono text-xs">{job.id}</span>
                  <span
                    className={`text-white text-xs px-2 py-0.5 rounded ${challengeGenerationStatusBadgeClass(job.status)}`}
                  >
                    {job.status}
                  </span>
                </div>
                <div className="text-gray-700">
                  Zone: {zoneNameById.get(job.zoneId) ?? job.zoneId}
                </div>
                {job.pointOfInterestId ? (
                  <div className="text-gray-700">
                    POI:{' '}
                    {allPointOfInterestNamesById.get(job.pointOfInterestId) ??
                      job.pointOfInterestId}
                  </div>
                ) : null}
                <div className="text-gray-700">
                  Created: {job.createdCount} / {job.count}
                </div>
                <div className="text-gray-600 text-xs">
                  Queued: {formatDate(job.createdAt)}
                </div>
                {job.errorMessage ? (
                  <div className="text-red-600 text-xs mt-1">
                    {job.errorMessage}
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="overflow-auto border rounded-md">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-100">
            <tr>
              <th className="text-left p-2 border-b w-10">Select</th>
              <th className="text-left p-2 border-b">Question</th>
              <th className="text-left p-2 border-b">Zone</th>
              <th className="text-left p-2 border-b">Submission</th>
              <th className="text-left p-2 border-b">Difficulty</th>
              <th className="text-left p-2 border-b">Reward</th>
              <th className="text-left p-2 border-b">Location</th>
              <th className="text-left p-2 border-b">Image</th>
              <th className="text-left p-2 border-b">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredRecords.length === 0 ? (
              <tr>
                <td className="p-3 text-gray-500" colSpan={9}>
                  No challenges found.
                </td>
              </tr>
            ) : (
              filteredRecords.map((record) => (
                <tr key={record.id} className="odd:bg-white even:bg-gray-50">
                  <td className="p-2 border-b align-top">
                    <input
                      type="checkbox"
                      className="h-4 w-4"
                      checked={selectedChallengeIdSet.has(record.id)}
                      disabled={bulkDeletingChallenges}
                      onChange={() => toggleChallengeSelection(record.id)}
                    />
                  </td>
                  <td className="p-2 border-b align-top max-w-md">
                    <div className="font-medium">{record.question}</div>
                    {record.description ? (
                      <div className="text-xs text-gray-700 mt-1 max-w-md whitespace-pre-wrap">
                        {record.description}
                      </div>
                    ) : null}
                    <div className="text-xs text-gray-500 font-mono mt-1">
                      {record.id}
                    </div>
                    {record.statTags?.length ? (
                      <div className="text-xs text-gray-600 mt-1">
                        Stats: {record.statTags.join(', ')}
                      </div>
                    ) : null}
                  </td>
                  <td className="p-2 border-b align-top">
                    {zoneNameById.get(record.zoneId) ?? record.zoneId}
                  </td>
                  <td className="p-2 border-b align-top">
                    {record.submissionType}
                  </td>
                  <td className="p-2 border-b align-top">
                    {record.difficulty}
                    {record.scaleWithUserLevel ? (
                      <div className="text-xs text-indigo-700">
                        scales with level
                      </div>
                    ) : null}
                    {record.recurrenceFrequency ? (
                      <div className="text-xs text-indigo-700">
                        repeats {record.recurrenceFrequency}
                      </div>
                    ) : null}
                  </td>
                  <td className="p-2 border-b align-top">
                    {(record.rewardMode ?? 'random') === 'random' ? (
                      <div>
                        <div className="text-xs font-medium text-indigo-700">
                          Random{' '}
                          {(record.randomRewardSize ?? 'small').toUpperCase()}
                        </div>
                      </div>
                    ) : (
                      <div>
                        XP {record.rewardExperience ?? 0} · Gold {record.reward}
                        {record.materialRewards &&
                        record.materialRewards.length > 0 ? (
                          <div className="text-xs text-gray-600">
                            {summarizeMaterialRewards(record.materialRewards)}
                          </div>
                        ) : null}
                        {record.inventoryItemId ? (
                          <div className="text-xs text-gray-600">
                            Item #{record.inventoryItemId}
                          </div>
                        ) : null}
                      </div>
                    )}
                  </td>
                  <td className="p-2 border-b align-top">
                    {record.polygonPoints && record.polygonPoints.length >= 3
                      ? `Polygon (${record.polygonPoints.length} points)`
                      : record.pointOfInterestId
                        ? `POI: ${
                            allPointOfInterestNamesById.get(
                              record.pointOfInterestId
                            ) ?? record.pointOfInterestId
                          }`
                        : Number.isFinite(record.latitude) &&
                            Number.isFinite(record.longitude)
                          ? `${record.latitude.toFixed(5)}, ${record.longitude.toFixed(5)}`
                          : 'n/a'}
                  </td>
                  <td className="p-2 border-b align-top">
                    {record.thumbnailUrl || record.imageUrl ? (
                      <button
                        type="button"
                        className="inline-flex"
                        onClick={() => openImagePreview(record)}
                        title="Open image preview"
                      >
                        <img
                          src={record.thumbnailUrl || record.imageUrl}
                          alt="Challenge"
                          className="w-12 h-12 object-cover rounded border"
                        />
                      </button>
                    ) : (
                      <span className="text-xs text-gray-500">No image</span>
                    )}
                  </td>
                  <td className="p-2 border-b align-top">
                    <div className="flex flex-col gap-2">
                      <button
                        type="button"
                        className="bg-indigo-600 text-white px-2 py-1 rounded-md disabled:opacity-60"
                        onClick={() => void handleGenerateImage(record)}
                        disabled={generatingChallengeId !== null}
                      >
                        {generatingChallengeId === record.id
                          ? 'Generating...'
                          : 'Generate Image'}
                      </button>
                      <button
                        type="button"
                        className="bg-blue-600 text-white px-2 py-1 rounded-md"
                        onClick={() => openEdit(record)}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="bg-red-600 text-white px-2 py-1 rounded-md"
                        onClick={() => void deleteChallenge(record)}
                        disabled={bulkDeletingChallenges}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <PaginationControls
        page={page}
        pageSize={challengeListPageSize}
        total={total}
        label="challenges"
        onPageChange={setPage}
      />

      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-4xl max-h-[90vh] overflow-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">
                {editingChallenge ? 'Edit Challenge' : 'Create Challenge'}
              </h2>
              <button
                type="button"
                className="text-gray-600 hover:text-black"
                onClick={closeModal}
              >
                Close
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <label className="text-sm">
                Zone
                <select
                  className="w-full border rounded-md p-2"
                  value={form.zoneId}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      zoneId: event.target.value,
                      pointOfInterestId: '',
                    }))
                  }
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
                Submission Type
                <select
                  className="w-full border rounded-md p-2"
                  value={form.submissionType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      submissionType: event.target.value as
                        | 'photo'
                        | 'text'
                        | 'video',
                    }))
                  }
                >
                  <option value="photo">photo</option>
                  <option value="text">text</option>
                  <option value="video">video</option>
                </select>
              </label>

              <label className="text-sm md:col-span-2">
                Question
                <textarea
                  className="w-full border rounded-md p-2 min-h-[100px]"
                  value={form.question}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      question: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm md:col-span-2">
                Description (Flavor)
                <textarea
                  className="w-full border rounded-md p-2 min-h-[120px]"
                  placeholder="Atmosphere, subject details, setting notes, tone, visual motifs."
                  value={form.description}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      description: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm">
                Difficulty
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                  value={form.difficulty}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      difficulty: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={form.scaleWithUserLevel}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      scaleWithUserLevel: event.target.checked,
                    }))
                  }
                />
                Scale difficulty with user level
              </label>

              <label className="text-sm">
                Recurrence
                <select
                  className="w-full border rounded-md p-2"
                  value={form.recurrenceFrequency}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: event.target.value,
                    }))
                  }
                >
                  {recurrenceOptions.map((option) => (
                    <option key={option.value || 'none'} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <label className="text-sm">
                Reward Mode
                <select
                  className="w-full border rounded-md p-2"
                  value={form.rewardMode}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardMode: event.target.value as 'explicit' | 'random',
                    }))
                  }
                >
                  <option value="random">Random Reward</option>
                  <option value="explicit">Explicit Reward</option>
                </select>
              </label>

              <label className="text-sm">
                Random Reward Size
                <select
                  className="w-full border rounded-md p-2"
                  value={form.randomRewardSize}
                  disabled={form.rewardMode !== 'random'}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      randomRewardSize: event.target.value as
                        | 'small'
                        | 'medium'
                        | 'large',
                    }))
                  }
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </label>

              <label className="text-sm">
                Reward Experience
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                  value={form.rewardExperience}
                  disabled={form.rewardMode !== 'explicit'}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardExperience: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm">
                Reward Gold
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                  value={form.reward}
                  disabled={form.rewardMode !== 'explicit'}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, reward: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Reward Item (Optional)
                <select
                  className="w-full border rounded-md p-2"
                  value={form.inventoryItemId}
                  disabled={form.rewardMode !== 'explicit'}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      inventoryItemId: event.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  {inventoryItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} (#{item.id})
                    </option>
                  ))}
                </select>
              </label>

              {form.rewardMode === 'random' ? (
                <div className="text-xs text-gray-500 md:col-span-2">
                  Random rewards ignore explicit XP, gold, material, and item
                  fields.
                </div>
              ) : null}

              <div className="md:col-span-2">
                <MaterialRewardsEditor
                  value={form.materialRewards}
                  onChange={(materialRewards) =>
                    setForm((prev) => ({ ...prev, materialRewards }))
                  }
                  disabled={form.rewardMode !== 'explicit'}
                />
              </div>

              <label className="text-sm">
                Proficiency (Optional)
                <input
                  className="w-full border rounded-md p-2"
                  value={form.proficiency}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      proficiency: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm md:col-span-2">
                Stat Tags (comma separated)
                <input
                  className="w-full border rounded-md p-2"
                  placeholder="strength, dexterity"
                  value={form.statTags}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      statTags: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="text-sm md:col-span-2">
                Location Mode
                <select
                  className="w-full border rounded-md p-2"
                  value={form.locationMode}
                  onChange={(event) => {
                    const nextMode = event.target.value as
                      | 'poi'
                      | 'coordinates'
                      | 'polygon';
                    setForm((prev) => ({
                      ...prev,
                      locationMode: nextMode,
                      pointOfInterestId:
                        nextMode === 'poi' ? prev.pointOfInterestId : '',
                      polygonPoints:
                        nextMode === 'polygon' ? prev.polygonPoints : '',
                    }));
                  }}
                >
                  <option value="coordinates">Coordinates</option>
                  <option value="poi">Point of Interest</option>
                  <option value="polygon">Polygon Area</option>
                </select>
              </label>

              {form.locationMode === 'poi' ? (
                <label className="text-sm md:col-span-2">
                  Point of Interest
                  <select
                    className="w-full border rounded-md p-2"
                    value={form.pointOfInterestId}
                    onChange={(event) => {
                      const nextPointOfInterestId = event.target.value;
                      if (!nextPointOfInterestId) {
                        setForm((prev) => ({ ...prev, pointOfInterestId: '' }));
                        return;
                      }
                      const selectedPoint = pointsOfInterestForFormZone.find(
                        (point) => point.id === nextPointOfInterestId
                      );
                      setForm((prev) => ({
                        ...prev,
                        pointOfInterestId: nextPointOfInterestId,
                        latitude:
                          selectedPoint?.latitude.toFixed(6) ?? prev.latitude,
                        longitude:
                          selectedPoint?.longitude.toFixed(6) ?? prev.longitude,
                      }));
                    }}
                  >
                    <option value="">Select point of interest</option>
                    {pointsOfInterestForFormZone.map((point) => (
                      <option key={point.id} value={point.id}>
                        {point.name}
                      </option>
                    ))}
                  </select>
                  {pointOfInterestLoadingByZone[form.zoneId] ? (
                    <div className="text-xs text-gray-500 mt-1">
                      Loading points of interest...
                    </div>
                  ) : null}
                </label>
              ) : null}

              {form.locationMode === 'coordinates' ? (
                <>
                  <label className="text-sm">
                    Latitude
                    <input
                      className="w-full border rounded-md p-2"
                      type="number"
                      step="any"
                      value={form.latitude}
                      onChange={(event) =>
                        setForm((prev) => ({
                          ...prev,
                          pointOfInterestId: '',
                          latitude: event.target.value,
                        }))
                      }
                    />
                  </label>

                  <label className="text-sm">
                    Longitude
                    <input
                      className="w-full border rounded-md p-2"
                      type="number"
                      step="any"
                      value={form.longitude}
                      onChange={(event) =>
                        setForm((prev) => ({
                          ...prev,
                          pointOfInterestId: '',
                          longitude: event.target.value,
                        }))
                      }
                    />
                  </label>
                </>
              ) : null}

              {form.locationMode === 'polygon' ? (
                <label className="text-sm md:col-span-2">
                  Polygon Points
                  <textarea
                    className="w-full border rounded-md p-2"
                    rows={4}
                    placeholder="Click on the map to draw the polygon"
                    value={form.polygonPoints}
                    readOnly
                  />
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-gray-600">
                    <span className="rounded-md bg-blue-50 px-2 py-1 text-blue-700">
                      Click on the map to add polygon points.
                    </span>
                    <button
                      type="button"
                      className="rounded-md border border-gray-300 bg-white px-2 py-1 text-gray-700 hover:bg-gray-50"
                      onClick={() => {
                        setPolygonDraftPoints([]);
                        setForm((prev) => ({ ...prev, polygonPoints: '' }));
                      }}
                    >
                      Clear polygon
                    </button>
                    <button
                      type="button"
                      className="rounded-md border border-gray-300 bg-white px-2 py-1 text-gray-700 hover:bg-gray-50"
                      onClick={() => {
                        setPolygonDraftPoints((prev) => {
                          if (prev.length === 0) return prev;
                          const next = prev.slice(0, -1);
                          setForm((formPrev) => ({
                            ...formPrev,
                            polygonPoints: JSON.stringify(next),
                          }));
                          return next;
                        });
                      }}
                    >
                      Undo last point
                    </button>
                    <span className="rounded-md bg-slate-100 px-2 py-1 text-slate-700">
                      {polygonDraftPoints.length} point
                      {polygonDraftPoints.length === 1 ? '' : 's'}
                    </span>
                  </div>
                </label>
              ) : null}
            </div>

            {isCoordinateMode ? (
              <div className="mt-3 mb-3">
                <button
                  type="button"
                  className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                  onClick={handleUseCurrentLocation}
                  disabled={geoLoading}
                >
                  {geoLoading ? 'Locating...' : 'Use Current Browser Location'}
                </button>
              </div>
            ) : null}

            {usesInteractiveMap ? (
              <div className="mb-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm">
                    {isPolygonMode
                      ? 'Polygon Draft Map'
                      : 'Map Location Picker'}
                  </span>
                  <span className="text-xs text-gray-500">
                    {isPolygonMode
                      ? 'Click map to add polygon points'
                      : 'Click map to set latitude/longitude'}
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
            ) : null}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-5">
              <label className="text-sm">
                Image URL (Optional)
                <input
                  className="w-full border rounded-md p-2"
                  value={form.imageUrl}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      imageUrl: event.target.value,
                    }))
                  }
                />
              </label>
              <label className="text-sm">
                Thumbnail URL (Optional)
                <input
                  className="w-full border rounded-md p-2"
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

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="bg-gray-200 px-4 py-2 rounded-md"
                onClick={closeModal}
              >
                Cancel
              </button>
              <button
                type="button"
                className="bg-blue-600 text-white px-4 py-2 rounded-md"
                onClick={() => void save()}
              >
                Save Challenge
              </button>
            </div>
          </div>
        </div>
      )}

      {imagePreview && (
        <div className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-6">
          <div className="relative max-w-4xl w-full">
            <button
              type="button"
              className="absolute -top-10 right-0 bg-white text-black px-3 py-1 rounded-md"
              onClick={closeImagePreview}
            >
              Close
            </button>
            <img
              src={imagePreview.url}
              alt={imagePreview.alt}
              className="w-full max-h-[80vh] object-contain rounded-lg bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
};
