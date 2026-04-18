import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI, useMediaContext, useTagContext, useZoneContext } from '@poltergeist/contexts';
import { useCandidates, usePointOfInterestGroups } from '@poltergeist/hooks';
import {
  Candidate,
  PointOfInterest as PointOfInterestType,
  Tag,
  ZoneGenre,
} from '@poltergeist/types';

type PointOfInterestImport = {
  id: string;
  placeId: string;
  zoneId: string;
  genreId: string;
  genre?: ZoneGenre | null;
  status: string;
  errorMessage?: string | null;
  pointOfInterestId?: string | null;
  createdAt: string;
  updatedAt: string;
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

type PoiMarkerCategoryIconResponse = StaticThumbnailResponse & {
  category: string;
  label: string;
  defaultPrompt: string;
};

type PoiMarkerCategoryIconState = PoiMarkerCategoryIconResponse & {
  prompt: string;
  busy: boolean;
  statusLoading: boolean;
  message: string | null;
  error: string | null;
  previewNonce: number;
};

const flattenTags = (tagGroups: { tags: Tag[] }[]): Tag[] => {
  return tagGroups.flatMap(group => group.tags);
};

const defaultPoiUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered point of interest. Enigmatic landmark silhouette with cartographer glyph motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

const staticStatusClassName = (status?: string) => {
  const normalized = (status || '').trim().toLowerCase();
  if (normalized === 'completed') return 'bg-emerald-600';
  if (normalized === 'queued' || normalized === 'in_progress')
    return 'bg-indigo-600';
  if (normalized === 'failed' || normalized === 'missing')
    return 'bg-red-600';
  return 'bg-gray-500';
};

const formatDate = (value?: string | null) => {
  if (!value) return '—';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const extractApiErrorMessage = (
  error: unknown,
  fallback: string
): string => {
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    typeof (error as { response?: unknown }).response === 'object'
  ) {
    const response = (error as { response?: { data?: unknown } }).response;
    const data = response?.data;
    if (typeof data === 'object' && data !== null) {
      const maybeMessage = (data as { error?: unknown; message?: unknown }).error;
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

export const PointOfInterest = () => {
  const { apiClient } = useAPI();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { zones } = useZoneContext();
  const { tagGroups } = useTagContext();
  const { pointOfInterestGroups } = usePointOfInterestGroups();

  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterestType[]>([]);
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [nameQuery, setNameQuery] = useState('');
  const [selectedZoneId, setSelectedZoneId] = useState('');
  const [selectedGenreId, setSelectedGenreId] = useState('all');
  const [selectedTagIds, setSelectedTagIds] = useState<Set<string>>(new Set());
  const [filtersOpen, setFiltersOpen] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [importError, setImportError] = useState<string | null>(null);
  const [poiUndiscoveredBusy, setPoiUndiscoveredBusy] = useState(false);
  const [poiUndiscoveredStatusLoading, setPoiUndiscoveredStatusLoading] =
    useState(false);
  const [poiUndiscoveredError, setPoiUndiscoveredError] = useState<string | null>(
    null
  );
  const [poiUndiscoveredMessage, setPoiUndiscoveredMessage] = useState<
    string | null
  >(null);
  const [poiUndiscoveredUrl, setPoiUndiscoveredUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png'
  );
  const [poiUndiscoveredStatus, setPoiUndiscoveredStatus] =
    useState('unknown');
  const [poiUndiscoveredExists, setPoiUndiscoveredExists] = useState(false);
  const [poiUndiscoveredRequestedAt, setPoiUndiscoveredRequestedAt] = useState<
    string | null
  >(null);
  const [poiUndiscoveredLastModified, setPoiUndiscoveredLastModified] = useState<
    string | null
  >(null);
  const [poiUndiscoveredPreviewNonce, setPoiUndiscoveredPreviewNonce] =
    useState(Date.now());
  const [poiUndiscoveredPrompt, setPoiUndiscoveredPrompt] = useState(
    defaultPoiUndiscoveredIconPrompt
  );
  const [poiMarkerCategoryOrder, setPoiMarkerCategoryOrder] = useState<string[]>(
    []
  );
  const [poiMarkerCategoryStates, setPoiMarkerCategoryStates] = useState<
    Record<string, PoiMarkerCategoryIconState>
  >({});
  const [poiMarkerCategoriesLoading, setPoiMarkerCategoriesLoading] =
    useState(false);
  const [poiMarkerCategoriesError, setPoiMarkerCategoriesError] = useState<
    string | null
  >(null);
  const [createForm, setCreateForm] = useState({
    genreId: '',
    name: '',
    description: '',
    clue: '',
    lat: '',
    lng: '',
    imageUrl: '',
    unlockTier: '',
    groupId: '',
  });
  const [createImageFile, setCreateImageFile] = useState<File | null>(null);
  const [createImagePreview, setCreateImagePreview] = useState<string | null>(null);

  const [importQuery, setImportQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(null);
  const [importZoneId, setImportZoneId] = useState('');
  const [importGenreId, setImportGenreId] = useState('');
  const { candidates } = useCandidates(importQuery);
  const [importJobs, setImportJobs] = useState<PointOfInterestImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const [importToasts, setImportToasts] = useState<string[]>([]);
  const [, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const [selectedPointIds, setSelectedPointIds] = useState<Set<string>>(new Set());
  const [deletingPointId, setDeletingPointId] = useState<string | null>(null);
  const [bulkDeletingPoints, setBulkDeletingPoints] = useState(false);

  const allTags = useMemo(() => flattenTags(tagGroups), [tagGroups]);
  const defaultGenreId = useMemo(() => {
    const fantasyGenre = genres.find(
      (genre) => genre.name.trim().toLowerCase() === 'fantasy'
    );
    return fantasyGenre?.id ?? genres[0]?.id ?? '';
  }, [genres]);
  const genreNameById = useMemo(
    () =>
      new Map(
        genres.map((genre) => [genre.id, genre.name])
      ),
    [genres]
  );
  const orderedPoiMarkerCategoryStates = useMemo(
    () =>
      poiMarkerCategoryOrder
        .map((category) => poiMarkerCategoryStates[category])
        .filter(
          (state): state is PoiMarkerCategoryIconState => state !== undefined
        ),
    [poiMarkerCategoryOrder, poiMarkerCategoryStates]
  );
  const fetchPointsOfInterest = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const endpoint = selectedZoneId
        ? `/sonar/zones/${selectedZoneId}/pointsOfInterest`
        : '/sonar/pointsOfInterest';
      const response = await apiClient.get<PointOfInterestType[]>(endpoint);
      setPointsOfInterest(response);
    } catch (err) {
      console.error('Error fetching points of interest:', err);
      setError('Failed to load points of interest');
    } finally {
      setLoading(false);
    }
  }, [apiClient, selectedZoneId]);

  useEffect(() => {
    fetchPointsOfInterest();
  }, [fetchPointsOfInterest]);

  useEffect(() => {
    const fetchGenres = async () => {
      try {
        const response = await apiClient.get<ZoneGenre[]>(
          '/sonar/zone-genres?includeInactive=true'
        );
        setGenres(response);
      } catch (err) {
        console.error('Error fetching genres:', err);
      }
    };

    void fetchGenres();
  }, [apiClient]);

  useEffect(() => {
    if (!defaultGenreId) {
      return;
    }
    setCreateForm((prev) =>
      prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
    );
    setImportGenreId((prev) => prev || defaultGenreId);
  }, [defaultGenreId]);

  const toggleTag = (tagId: string) => {
    setSelectedTagIds(prev => {
      const next = new Set(prev);
      if (next.has(tagId)) {
        next.delete(tagId);
      } else {
        next.add(tagId);
      }
      return next;
    });
  };

  const clearTags = () => setSelectedTagIds(new Set());

  const filteredPoints = useMemo(() => {
    const query = nameQuery.trim().toLowerCase();
    const selectedTags = Array.from(selectedTagIds);

    return pointsOfInterest.filter(point => {
      const matchesName = query.length === 0
        || point.name.toLowerCase().includes(query)
        || (point.originalName && point.originalName.toLowerCase().includes(query));

      if (!matchesName) {
        return false;
      }

      const pointGenreId = point.genreId ?? point.genre?.id ?? '';
      if (selectedGenreId !== 'all' && pointGenreId !== selectedGenreId) {
        return false;
      }

      if (selectedTags.length === 0) {
        return true;
      }

      const pointTagIds = point.tags?.map(tag => tag.id) ?? [];
      return selectedTags.every(tagId => pointTagIds.includes(tagId));
    });
  }, [pointsOfInterest, nameQuery, selectedGenreId, selectedTagIds]);

  const allFilteredPointsSelected = useMemo(() => {
    if (filteredPoints.length === 0) return false;
    return filteredPoints.every(point => selectedPointIds.has(point.id));
  }, [filteredPoints, selectedPointIds]);

  const resetCreateForm = () => {
    setCreateForm({
      genreId: defaultGenreId,
      name: '',
      description: '',
      clue: '',
      lat: '',
      lng: '',
      imageUrl: '',
      unlockTier: '',
      groupId: '',
    });
    setCreateImageFile(null);
    setCreateImagePreview(null);
    setCreateError(null);
  };

  const resetImportForm = () => {
    setImportQuery('');
    setSelectedCandidate(null);
    setImportZoneId('');
    setImportGenreId(defaultGenreId);
    setImportError(null);
  };

  const refreshPoiUndiscoveredIconStatus = React.useCallback(
    async (showMessage = false) => {
      try {
        setPoiUndiscoveredStatusLoading(true);
        setPoiUndiscoveredError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/poi-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setPoiUndiscoveredUrl(url);
        }
        setPoiUndiscoveredStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setPoiUndiscoveredExists(Boolean(response?.exists));
        setPoiUndiscoveredRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setPoiUndiscoveredLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setPoiUndiscoveredPreviewNonce(Date.now());
        if (showMessage) {
          setPoiUndiscoveredMessage('Undiscovered POI icon status refreshed.');
        }
      } catch (err) {
        console.error('Failed to load undiscovered POI icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load undiscovered POI icon status.';
        setPoiUndiscoveredError(message);
      } finally {
        setPoiUndiscoveredStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGeneratePoiUndiscoveredIcon = React.useCallback(async () => {
    const prompt = poiUndiscoveredPrompt.trim();
    if (!prompt) {
      setPoiUndiscoveredError('Prompt is required.');
      return;
    }
    try {
      setPoiUndiscoveredBusy(true);
      setPoiUndiscoveredError(null);
      setPoiUndiscoveredMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/poi-undiscovered',
        { prompt }
      );
      setPoiUndiscoveredMessage('Undiscovered POI icon queued for generation.');
      await refreshPoiUndiscoveredIconStatus();
    } catch (err) {
      console.error('Failed to generate undiscovered POI icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate undiscovered POI icon.';
      setPoiUndiscoveredError(message);
    } finally {
      setPoiUndiscoveredBusy(false);
    }
  }, [apiClient, poiUndiscoveredPrompt, refreshPoiUndiscoveredIconStatus]);

  const handleDeletePoiUndiscoveredIcon = React.useCallback(async () => {
    try {
      setPoiUndiscoveredBusy(true);
      setPoiUndiscoveredError(null);
      setPoiUndiscoveredMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/poi-undiscovered'
      );
      setPoiUndiscoveredMessage('Undiscovered POI icon deleted.');
      await refreshPoiUndiscoveredIconStatus();
    } catch (err) {
      console.error('Failed to delete undiscovered POI icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete undiscovered POI icon.';
      setPoiUndiscoveredError(message);
    } finally {
      setPoiUndiscoveredBusy(false);
    }
  }, [apiClient, refreshPoiUndiscoveredIconStatus]);

  const applyPoiMarkerCategorySnapshots = React.useCallback(
    (snapshots: PoiMarkerCategoryIconResponse[]) => {
      setPoiMarkerCategoryOrder(snapshots.map((snapshot) => snapshot.category));
      setPoiMarkerCategoryStates((prev) => {
        const next: Record<string, PoiMarkerCategoryIconState> = {};
        snapshots.forEach((snapshot) => {
          const existing = prev[snapshot.category];
          const changed =
            existing?.thumbnailUrl !== snapshot.thumbnailUrl ||
            existing?.status !== snapshot.status ||
            existing?.exists !== Boolean(snapshot.exists) ||
            existing?.requestedAt !== snapshot.requestedAt ||
            existing?.lastModified !== snapshot.lastModified;
          next[snapshot.category] = {
            ...snapshot,
            prompt: existing?.prompt ?? snapshot.defaultPrompt,
            busy: existing?.busy ?? false,
            statusLoading: false,
            message: existing?.message ?? null,
            error: existing?.error ?? null,
            previewNonce:
              existing && !changed ? existing.previewNonce : Date.now(),
          };
        });
        return next;
      });
    },
    []
  );

  const fetchPoiMarkerCategoryIcons = React.useCallback(async () => {
    try {
      setPoiMarkerCategoriesLoading(true);
      setPoiMarkerCategoriesError(null);
      const response = await apiClient.get<PoiMarkerCategoryIconResponse[]>(
        '/sonar/admin/thumbnails/poi-marker-categories'
      );
      applyPoiMarkerCategorySnapshots(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load POI marker category icons', err);
      setPoiMarkerCategoriesError(
        extractApiErrorMessage(
          err,
          'Failed to load discovered POI category icons.'
        )
      );
    } finally {
      setPoiMarkerCategoriesLoading(false);
    }
  }, [apiClient, applyPoiMarkerCategorySnapshots]);

  const refreshPoiMarkerCategoryIconStatus = React.useCallback(
    async (category: string, showMessage = false) => {
      setPoiMarkerCategoryStates((prev) => {
        const existing = prev[category];
        if (!existing) return prev;
        return {
          ...prev,
          [category]: {
            ...existing,
            statusLoading: true,
            error: null,
          },
        };
      });
      try {
        const response = await apiClient.get<PoiMarkerCategoryIconResponse>(
          `/sonar/admin/thumbnails/poi-marker-categories/${category}/status`
        );
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) {
            return prev;
          }
          return {
            ...prev,
            [category]: {
              ...existing,
              ...response,
              prompt: existing.prompt || response.defaultPrompt,
              busy: false,
              statusLoading: false,
              message: showMessage
                ? `${response.label} icon status refreshed.`
                : existing.message,
              error: null,
              previewNonce: Date.now(),
            },
          };
        });
      } catch (err) {
        console.error(
          `Failed to refresh POI marker category icon status for ${category}`,
          err
        );
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) {
            return prev;
          }
          return {
            ...prev,
            [category]: {
              ...existing,
              statusLoading: false,
              error: extractApiErrorMessage(
                err,
                `Failed to load ${existing.label} icon status.`
              ),
            },
          };
        });
      }
    },
    [apiClient]
  );

  const handlePoiMarkerCategoryPromptChange = React.useCallback(
    (category: string, value: string) => {
      setPoiMarkerCategoryStates((prev) => {
        const existing = prev[category];
        if (!existing) return prev;
        return {
          ...prev,
          [category]: {
            ...existing,
            prompt: value,
          },
        };
      });
    },
    []
  );

  const handleGeneratePoiMarkerCategoryIcon = React.useCallback(
    async (category: string) => {
      const current = poiMarkerCategoryStates[category];
      if (!current) return;
      const prompt = current.prompt.trim();
      if (!prompt) {
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              error: 'Prompt is required.',
            },
          };
        });
        return;
      }
      try {
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              busy: true,
              error: null,
              message: null,
            },
          };
        });
        await apiClient.post(
          `/sonar/admin/thumbnails/poi-marker-categories/${category}`,
          { prompt }
        );
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              message: `${current.label} icon queued for generation.`,
            },
          };
        });
        await refreshPoiMarkerCategoryIconStatus(category);
      } catch (err) {
        console.error(`Failed to generate ${current.label} icon`, err);
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              error: extractApiErrorMessage(
                err,
                `Failed to generate ${current.label} icon.`
              ),
            },
          };
        });
      } finally {
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              busy: false,
            },
          };
        });
      }
    },
    [apiClient, poiMarkerCategoryStates, refreshPoiMarkerCategoryIconStatus]
  );

  const handleDeletePoiMarkerCategoryIcon = React.useCallback(
    async (category: string) => {
      const current = poiMarkerCategoryStates[category];
      if (!current) return;
      try {
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              busy: true,
              error: null,
              message: null,
            },
          };
        });
        await apiClient.delete(
          `/sonar/admin/thumbnails/poi-marker-categories/${category}`
        );
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              message: `${current.label} icon deleted.`,
            },
          };
        });
        await refreshPoiMarkerCategoryIconStatus(category);
      } catch (err) {
        console.error(`Failed to delete ${current.label} icon`, err);
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              error: extractApiErrorMessage(
                err,
                `Failed to delete ${current.label} icon.`
              ),
            },
          };
        });
      } finally {
        setPoiMarkerCategoryStates((prev) => {
          const existing = prev[category];
          if (!existing) return prev;
          return {
            ...prev,
            [category]: {
              ...existing,
              busy: false,
            },
          };
        });
      }
    },
    [apiClient, poiMarkerCategoryStates, refreshPoiMarkerCategoryIconStatus]
  );

  useEffect(() => {
    void refreshPoiUndiscoveredIconStatus();
  }, [refreshPoiUndiscoveredIconStatus]);

  useEffect(() => {
    void fetchPoiMarkerCategoryIcons();
  }, [fetchPoiMarkerCategoryIcons]);

  useEffect(() => {
    if (
      poiUndiscoveredStatus !== 'queued' &&
      poiUndiscoveredStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshPoiUndiscoveredIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [poiUndiscoveredStatus, refreshPoiUndiscoveredIconStatus]);

  useEffect(() => {
    if (
      !orderedPoiMarkerCategoryStates.some(
        (state) =>
          state.status === 'queued' || state.status === 'in_progress'
      )
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void fetchPoiMarkerCategoryIcons();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [fetchPoiMarkerCategoryIcons, orderedPoiMarkerCategoryStates]);

  const handleCreateImageChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    setCreateImageFile(file);
    if (!file) {
      setCreateImagePreview(null);
      return;
    }
    const reader = new FileReader();
    reader.onloadend = () => {
      setCreateImagePreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const uploadCreateImage = async () => {
    if (!createImageFile) return null;
    const imageKey = `points-of-interest/${createImageFile.name.toLowerCase().replace(/\s+/g, '-')}`;
    const presignedUrl = await getPresignedUploadURL('crew-points-of-interest', imageKey);
    if (!presignedUrl) {
      throw new Error('Failed to get upload URL');
    }
    await uploadMedia(presignedUrl, createImageFile);
    return presignedUrl.split('?')[0];
  };

  const handleCreatePointOfInterest = async () => {
    setCreateError(null);
    if (!createForm.groupId) {
      setCreateError('Please select a point of interest group.');
      return;
    }
    if (!createForm.genreId) {
      setCreateError('Please select a genre.');
      return;
    }
    if (!createForm.name || !createForm.description || !createForm.clue) {
      setCreateError('Name, description, and clue are required.');
      return;
    }
    if (!createForm.lat || !createForm.lng) {
      setCreateError('Latitude and longitude are required.');
      return;
    }

    try {
      let imageUrl = createForm.imageUrl;
      if (createImageFile) {
        const uploadedUrl = await uploadCreateImage();
        if (uploadedUrl) {
          imageUrl = uploadedUrl;
        }
      }

      if (!imageUrl) {
        setCreateError('Please provide an image URL or upload an image.');
        return;
      }

      await apiClient.post(`/sonar/pointsOfInterest/group/${createForm.groupId}`, {
        genreId: createForm.genreId,
        name: createForm.name,
        description: createForm.description,
        latitude: createForm.lat,
        longitude: createForm.lng,
        imageUrl,
        clue: createForm.clue,
        unlockTier: createForm.unlockTier ? Number(createForm.unlockTier) : null,
      });

      setShowCreateModal(false);
      resetCreateForm();
      fetchPointsOfInterest();
    } catch (err) {
      console.error('Error creating point of interest:', err);
      setCreateError('Failed to create point of interest.');
    }
  };

  const handleImportPointOfInterest = async () => {
    setImportError(null);
    if (!selectedCandidate) {
      setImportError('Please select a Google Maps location.');
      return;
    }
    if (!importZoneId) {
      setImportError('Please select a zone.');
      return;
    }
    if (!importGenreId) {
      setImportError('Please select a genre.');
      return;
    }
    try {
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: selectedCandidate.place_id,
        zoneID: importZoneId,
        genreId: importGenreId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (err) {
      console.error('Error importing point of interest:', err);
      setImportError('Failed to import point of interest.');
    }
  };

  const handleRetryImport = async (
    placeId: string,
    zoneId: string,
    genreId: string
  ) => {
    try {
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: placeId,
        zoneID: zoneId,
        genreId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (err) {
      console.error('Error retrying import:', err);
      setImportError('Failed to retry import.');
    }
  };

  const fetchImportJobs = useCallback(async (zoneId?: string) => {
    try {
      const url = zoneId ? `/sonar/pointOfInterest/imports?zoneId=${zoneId}` : '/sonar/pointOfInterest/imports';
      const response = await apiClient.get<PointOfInterestImport[]>(url);
      setImportJobs(response);
      const hasPending = response.some((item) => item.status === 'queued' || item.status === 'in_progress');
      setImportPolling(hasPending);
      if (!hasPending) {
        fetchPointsOfInterest();
      }
    } catch (err) {
      console.error('Error fetching import status:', err);
    }
  }, [apiClient, fetchPointsOfInterest]);

  useEffect(() => {
    if (!showImportModal) return;
    fetchImportJobs(importZoneId || undefined);
  }, [fetchImportJobs, importZoneId, showImportModal]);

  useEffect(() => {
    if (!importPolling) return;
    const interval = setInterval(() => {
      fetchImportJobs(importZoneId || undefined);
    }, 3000);
    return () => clearInterval(interval);
  }, [fetchImportJobs, importPolling, importZoneId]);

  useEffect(() => {
    if (importJobs.length === 0) return;
    const completed = importJobs.filter((job) => job.status === 'completed' && job.pointOfInterestId);
    if (completed.length === 0) return;

    setNotifiedImportIds((prev) => {
      const next = new Set(prev);
      let hasNew = false;
      completed.forEach((job) => {
        if (!next.has(job.id)) {
          next.add(job.id);
          hasNew = true;
          setImportToasts((existing) => [`Import complete: ${job.placeId}`, ...existing].slice(0, 3));
        }
      });
      return hasNew ? next : prev;
    });
  }, [importJobs]);

  const togglePointSelection = (pointId: string) => {
    setSelectedPointIds((prev) => {
      const next = new Set(prev);
      if (next.has(pointId)) {
        next.delete(pointId);
      } else {
        next.add(pointId);
      }
      return next;
    });
  };

  const toggleSelectVisiblePoints = () => {
    if (filteredPoints.length === 0) return;
    setSelectedPointIds((prev) => {
      const next = new Set(prev);
      if (allFilteredPointsSelected) {
        filteredPoints.forEach((point) => next.delete(point.id));
      } else {
        filteredPoints.forEach((point) => next.add(point.id));
      }
      return next;
    });
  };

  const clearPointSelection = () => {
    setSelectedPointIds(new Set());
  };

  const handleDeletePoint = async (point: PointOfInterestType) => {
    if (deletingPointId || bulkDeletingPoints) return;

    const pointLabel = point.name?.trim() || point.id;
    if (
      !window.confirm(
        `Delete point of interest "${pointLabel}"? This cannot be undone.`
      )
    ) {
      return;
    }

    try {
      setDeletingPointId(point.id);
      await apiClient.delete(`/sonar/pointsOfInterest/${point.id}`);
      setPointsOfInterest((prev) => prev.filter((poi) => poi.id !== point.id));
      setSelectedPointIds((prev) => {
        if (!prev.has(point.id)) return prev;
        const next = new Set(prev);
        next.delete(point.id);
        return next;
      });
    } catch (err) {
      console.error('Error deleting point of interest:', err);
      alert('Failed to delete point of interest.');
    } finally {
      setDeletingPointId(null);
    }
  };

  const handleBulkDeletePoints = async () => {
    if (
      bulkDeletingPoints ||
      selectedPointIds.size === 0 ||
      deletingPointId !== null
    ) {
      return;
    }

    const selectedIds = Array.from(selectedPointIds);
    const selectedNames = pointsOfInterest
      .filter((point) => selectedPointIds.has(point.id))
      .map((point) => point.name || point.id);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected point of interest (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected points of interest${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingPoints(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((pointId) =>
          apiClient.delete(`/sonar/pointsOfInterest/${pointId}`)
        )
      );

      const deletedIds = new Set<string>();
      const failedIds: string[] = [];
      results.forEach((result, index) => {
        const pointId = selectedIds[index];
        if (result.status === 'fulfilled') {
          deletedIds.add(pointId);
        } else {
          console.error(`Failed to delete point of interest ${pointId}`, result.reason);
          failedIds.push(pointId);
        }
      });

      if (deletedIds.size > 0) {
        setPointsOfInterest((prev) =>
          prev.filter((point) => !deletedIds.has(point.id))
        );
        setSelectedPointIds((prev) => {
          const next = new Set(prev);
          deletedIds.forEach((pointId) => next.delete(pointId));
          return next;
        });
      }

      if (failedIds.length > 0) {
        alert(
          `Deleted ${deletedIds.size} point${
            deletedIds.size === 1 ? '' : 's'
          }, but failed to delete ${failedIds.length}. Check console for details.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk delete points of interest', err);
      alert('Failed to delete selected points of interest.');
    } finally {
      setBulkDeletingPoints(false);
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      {importToasts.length > 0 && (
        <div className="fixed right-4 top-4 z-50 space-y-2">
          {importToasts.map((toast, index) => (
            <div
              key={`${toast}-${index}`}
              className="rounded-md bg-emerald-600 px-4 py-2 text-sm text-white shadow"
            >
              {toast}
            </div>
          ))}
        </div>
      )}
      <div className="flex flex-col gap-2 mb-6">
        <h1 className="text-3xl font-bold">Points of Interest</h1>
        <p className="text-gray-600">
          Search by name and filter by zone, genre, or tags.
        </p>
      </div>

      <div className="flex flex-wrap gap-3 mb-6">
        <button
          type="button"
          className="bg-blue-600 text-white px-4 py-2 rounded-md"
          onClick={() => {
            resetCreateForm();
            setShowCreateModal(true);
          }}
        >
          Create Point of Interest
        </button>
        <button
          type="button"
          className="bg-green-600 text-white px-4 py-2 rounded-md"
          onClick={() => {
            resetImportForm();
            setShowImportModal(true);
          }}
        >
          Import from Google Maps
        </button>
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">Undiscovered POI Icon</h2>
          <div className="flex gap-2">
            <button
              type="button"
              className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={() => void refreshPoiUndiscoveredIconStatus(true)}
              disabled={poiUndiscoveredStatusLoading}
            >
              {poiUndiscoveredStatusLoading ? 'Refreshing…' : 'Refresh Status'}
            </button>
            <button
              type="button"
              className="bg-indigo-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleGeneratePoiUndiscoveredIcon}
              disabled={poiUndiscoveredBusy || poiUndiscoveredStatusLoading}
            >
              {poiUndiscoveredBusy ? 'Working…' : 'Generate Icon'}
            </button>
            <button
              type="button"
              className="bg-red-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleDeletePoiUndiscoveredIcon}
              disabled={poiUndiscoveredBusy || poiUndiscoveredStatusLoading}
            >
              {poiUndiscoveredBusy ? 'Working…' : 'Delete Icon'}
            </button>
          </div>
        </div>
        <div className="mb-2">
          <span
            className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
              poiUndiscoveredStatus
            )}`}
          >
            {poiUndiscoveredStatus || 'unknown'}
          </span>
        </div>
        <div className="text-xs text-gray-600 break-all">
          URL: {poiUndiscoveredUrl}
        </div>
        <div className="text-xs text-gray-600 mt-1">
          Requested: {formatDate(poiUndiscoveredRequestedAt ?? undefined)}
          {' · '}
          Last updated: {formatDate(poiUndiscoveredLastModified ?? undefined)}
        </div>
        <label className="block text-sm mt-3">
          Generation Prompt
          <textarea
            className="w-full border rounded-md p-2 mt-1 min-h-[88px]"
            value={poiUndiscoveredPrompt}
            onChange={(event) => setPoiUndiscoveredPrompt(event.target.value)}
            placeholder="Prompt used to generate the undiscovered POI icon."
          />
        </label>
        {poiUndiscoveredExists ? (
          <div className="mt-3">
            <img
              src={`${poiUndiscoveredUrl}?v=${poiUndiscoveredPreviewNonce}`}
              alt="Undiscovered POI icon preview"
              className="w-24 h-24 object-cover border rounded-md bg-gray-50"
            />
          </div>
        ) : (
          <div className="text-xs text-gray-500 mt-2">
            No icon currently found at this URL.
          </div>
        )}
        {poiUndiscoveredMessage ? (
          <div className="text-sm text-emerald-700 mt-2">
            {poiUndiscoveredMessage}
          </div>
        ) : null}
        {poiUndiscoveredError ? (
          <div className="text-sm text-red-600 mt-2">
            {poiUndiscoveredError}
          </div>
        ) : null}
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3 mb-3">
          <div>
            <h2 className="text-lg font-semibold">
              Discovered POI Category Icons
            </h2>
            <p className="text-sm text-gray-600 mt-1">
              Shared map icons for discovered POIs by marker category. If a
              category image is missing, the game falls back to the in-app
              generated marker.
            </p>
          </div>
          <button
            type="button"
            className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
            onClick={() => void fetchPoiMarkerCategoryIcons()}
            disabled={poiMarkerCategoriesLoading}
          >
            {poiMarkerCategoriesLoading ? 'Refreshing…' : 'Refresh All'}
          </button>
        </div>

        {poiMarkerCategoriesError ? (
          <div className="text-sm text-red-600 mb-3">
            {poiMarkerCategoriesError}
          </div>
        ) : null}

        {orderedPoiMarkerCategoryStates.length === 0 ? (
          <div className="text-sm text-gray-500">
            {poiMarkerCategoriesLoading
              ? 'Loading category icons...'
              : 'No discovered POI category icons found.'}
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-2">
            {orderedPoiMarkerCategoryStates.map((state) => (
              <div
                key={state.category}
                className="rounded-md border border-gray-200 p-4"
              >
                <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
                  <div>
                    <h3 className="text-base font-semibold">{state.label}</h3>
                    <div className="text-xs text-gray-500">
                      Category: {state.category}
                    </div>
                  </div>
                  <span
                    className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
                      state.status
                    )}`}
                  >
                    {state.status || 'unknown'}
                  </span>
                </div>

                <div className="text-xs text-gray-600 break-all">
                  URL: {state.thumbnailUrl}
                </div>
                <div className="text-xs text-gray-600 mt-1">
                  Requested: {formatDate(state.requestedAt ?? undefined)}
                  {' · '}
                  Last updated: {formatDate(state.lastModified ?? undefined)}
                </div>

                <label className="block text-sm mt-3">
                  Generation Prompt
                  <textarea
                    className="w-full border rounded-md p-2 mt-1 min-h-[88px]"
                    value={state.prompt}
                    onChange={(event) =>
                      handlePoiMarkerCategoryPromptChange(
                        state.category,
                        event.target.value
                      )
                    }
                    placeholder={`Prompt used to generate the ${state.label.toLowerCase()} marker icon.`}
                  />
                </label>

                <div className="flex flex-wrap gap-2 mt-3">
                  <button
                    type="button"
                    className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
                    onClick={() =>
                      void refreshPoiMarkerCategoryIconStatus(
                        state.category,
                        true
                      )
                    }
                    disabled={state.statusLoading || state.busy}
                  >
                    {state.statusLoading ? 'Refreshing…' : 'Refresh Status'}
                  </button>
                  <button
                    type="button"
                    className="bg-indigo-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
                    onClick={() =>
                      void handleGeneratePoiMarkerCategoryIcon(state.category)
                    }
                    disabled={state.statusLoading || state.busy}
                  >
                    {state.busy ? 'Working…' : 'Generate Icon'}
                  </button>
                  <button
                    type="button"
                    className="bg-red-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
                    onClick={() =>
                      void handleDeletePoiMarkerCategoryIcon(state.category)
                    }
                    disabled={state.statusLoading || state.busy}
                  >
                    {state.busy ? 'Working…' : 'Delete Icon'}
                  </button>
                </div>

                {state.exists ? (
                  <div className="mt-3">
                    <img
                      src={`${state.thumbnailUrl}?v=${state.previewNonce}`}
                      alt={`${state.label} marker icon preview`}
                      className="w-24 h-24 object-cover border rounded-md bg-gray-50"
                    />
                  </div>
                ) : (
                  <div className="text-xs text-gray-500 mt-2">
                    No icon currently found at this URL.
                  </div>
                )}

                {state.message ? (
                  <div className="text-sm text-emerald-700 mt-2">
                    {state.message}
                  </div>
                ) : null}
                {state.error ? (
                  <div className="text-sm text-red-600 mt-2">
                    {state.error}
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="bg-white rounded-lg shadow-md p-4 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Filters</h2>
          <button
            type="button"
            onClick={() => setFiltersOpen((prev) => !prev)}
            className="text-sm text-blue-600 hover:underline"
          >
            {filtersOpen ? 'Hide filters' : 'Show filters'}
          </button>
        </div>

        {filtersOpen && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Search</label>
                <input
                  type="text"
                  placeholder="Search by name..."
                  value={nameQuery}
                  onChange={(e) => setNameQuery(e.target.value)}
                  className="border border-gray-300 rounded-md px-3 py-2"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Zone</label>
                <select
                  className="border border-gray-300 rounded-md px-3 py-2"
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                >
                  <option value="">All zones</option>
                  {zones.map(zone => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Genre</label>
                <select
                  className="border border-gray-300 rounded-md px-3 py-2"
                  value={selectedGenreId}
                  onChange={(e) => setSelectedGenreId(e.target.value)}
                >
                  <option value="all">All genres</option>
                  {genres.map((genre) => (
                    <option key={genre.id} value={genre.id}>
                      {genre.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Tags</label>
                <div className="flex flex-wrap gap-2">
                  {allTags.map(tag => {
                    const isSelected = selectedTagIds.has(tag.id);
                    return (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => toggleTag(tag.id)}
                        className={`px-3 py-1 rounded-full text-sm border transition-colors ${
                          isSelected
                            ? 'bg-blue-600 text-white border-blue-600'
                            : 'bg-gray-100 text-gray-700 border-gray-200 hover:bg-gray-200'
                        }`}
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                  {allTags.length === 0 && (
                    <span className="text-sm text-gray-500">No tags available</span>
                  )}
                </div>
                {selectedTagIds.size > 0 && (
                  <button
                    type="button"
                    onClick={clearTags}
                    className="text-sm text-blue-600 hover:underline w-fit"
                  >
                    Clear tag filters
                  </button>
                )}
              </div>
            </div>

            <div className="mt-4 text-sm text-gray-600">
              Showing {filteredPoints.length} of {pointsOfInterest.length} points
              {selectedTagIds.size > 0 && ' (matches all selected tags)'}
            </div>
          </>
        )}
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50 disabled:opacity-60"
          onClick={toggleSelectVisiblePoints}
          disabled={
            filteredPoints.length === 0 ||
            bulkDeletingPoints ||
            deletingPointId !== null
          }
        >
          {allFilteredPointsSelected ? 'Unselect Visible' : 'Select Visible'}
        </button>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50 disabled:opacity-60"
          onClick={clearPointSelection}
          disabled={selectedPointIds.size === 0 || bulkDeletingPoints}
        >
          Clear Selection
        </button>
        <button
          type="button"
          className="rounded-md bg-red-600 px-2 py-1 text-xs text-white disabled:opacity-60"
          onClick={handleBulkDeletePoints}
          disabled={
            selectedPointIds.size === 0 ||
            bulkDeletingPoints ||
            deletingPointId !== null
          }
        >
          {bulkDeletingPoints
            ? `Deleting ${selectedPointIds.size}...`
            : `Delete Selected (${selectedPointIds.size})`}
        </button>
      </div>

      {loading && (
        <div className="text-gray-600">Loading points of interest...</div>
      )}
      {error && (
        <div className="text-red-600">{error}</div>
      )}

      {!loading && !error && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredPoints.map(point => {
            const isDeleting = deletingPointId === point.id;

            return (
              <div key={point.id} className="bg-white rounded-lg shadow-md p-4 h-full">
                <div className="flex items-start justify-between gap-2 mb-3">
                  <div className="text-xs text-gray-500 break-all">{point.id}</div>
                  <div className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      className="h-4 w-4"
                      checked={selectedPointIds.has(point.id)}
                      disabled={bulkDeletingPoints || deletingPointId !== null}
                      onChange={() => togglePointSelection(point.id)}
                    />
                    <button
                      type="button"
                      className="rounded-md bg-red-600 px-2 py-1 text-xs text-white disabled:opacity-60"
                      onClick={() => void handleDeletePoint(point)}
                      disabled={bulkDeletingPoints || deletingPointId !== null}
                    >
                      {isDeleting ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </div>

                <Link to={`/points-of-interest/${point.id}`} className="block hover:opacity-90 transition-opacity">
                  {point.imageURL && (
                    <img
                      src={point.imageURL}
                      alt={point.name}
                      className="w-full h-40 object-cover rounded mb-3"
                    />
                  )}
                  <h2 className="text-xl font-semibold mb-1">{point.name}</h2>
                  <p className="text-xs font-medium uppercase tracking-wide text-blue-700 mb-2">
                    {point.genre?.name ??
                      genreNameById.get(point.genreId ?? '') ??
                      'Unknown genre'}
                  </p>
                  {point.originalName && (
                    <p className="text-sm text-gray-500 mb-2">{point.originalName}</p>
                  )}
                  <p className="text-sm text-gray-700 mb-2">{point.description}</p>
                  <div className="text-sm text-gray-600 space-y-1">
                    <div>Latitude: {point.lat}</div>
                    <div>Longitude: {point.lng}</div>
                    <div>Clue: {point.clue || 'None'}</div>
                    <div>
                      Tags: {point.tags?.length ? point.tags.map(tag => tag.name).join(', ') : 'None'}
                    </div>
                  </div>
                </Link>
                <div className="mt-3 text-right">
                  <Link
                    to={`/points-of-interest/${point.id}`}
                    className="text-sm text-blue-600 hover:underline"
                  >
                    Open Editor
                  </Link>
                </div>
              </div>
            );
          })}

          {filteredPoints.length === 0 && (
            <div className="col-span-full text-gray-600">
              No points of interest match these filters.
            </div>
          )}
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">Create Point of Interest</h2>

            {createError && (
              <div className="mb-4 text-red-600 text-sm">{createError}</div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Group</label>
                <select
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.groupId}
                  onChange={(e) => setCreateForm({ ...createForm, groupId: e.target.value })}
                >
                  <option value="">Select a group</option>
                  {(pointOfInterestGroups ?? []).map(group => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Genre</label>
                <select
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.genreId}
                  onChange={(e) =>
                    setCreateForm({ ...createForm, genreId: e.target.value })
                  }
                >
                  <option value="">Select a genre</option>
                  {genres.map((genre) => (
                    <option key={genre.id} value={genre.id}>
                      {genre.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.name}
                  onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Latitude</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.lat}
                  onChange={(e) => setCreateForm({ ...createForm, lat: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Longitude</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.lng}
                  onChange={(e) => setCreateForm({ ...createForm, lng: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Unlock Tier</label>
                <input
                  type="number"
                  min="1"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.unlockTier}
                  onChange={(e) => setCreateForm({ ...createForm, unlockTier: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Image URL</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.imageUrl}
                  onChange={(e) => setCreateForm({ ...createForm, imageUrl: e.target.value })}
                />
              </div>
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                rows={3}
                value={createForm.description}
                onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
              />
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Clue</label>
              <textarea
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                rows={2}
                value={createForm.clue}
                onChange={(e) => setCreateForm({ ...createForm, clue: e.target.value })}
              />
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Upload Image</label>
              <input type="file" accept="image/*" onChange={handleCreateImageChange} />
              {createImagePreview && (
                <img src={createImagePreview} alt="Preview" className="mt-2 h-32 w-full object-cover rounded" />
              )}
            </div>

            <div className="flex justify-end gap-2 mt-6">
              <button
                type="button"
                className="px-4 py-2 rounded-md border border-gray-300"
                onClick={() => {
                  setShowCreateModal(false);
                  resetCreateForm();
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-md bg-blue-600 text-white"
                onClick={handleCreatePointOfInterest}
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">Import Point of Interest</h2>

            {importError && (
              <div className="mb-4 text-red-600 text-sm">{importError}</div>
            )}

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Zone</label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={importZoneId}
                onChange={(e) => setImportZoneId(e.target.value)}
              >
                <option value="">Select a zone</option>
                {zones.map(zone => (
                  <option key={zone.id} value={zone.id}>
                    {zone.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Genre</label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={importGenreId}
                onChange={(e) => setImportGenreId(e.target.value)}
              >
                <option value="">Select a genre</option>
                {genres.map((genre) => (
                  <option key={genre.id} value={genre.id}>
                    {genre.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Search Google Maps</label>
              <input
                type="text"
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={importQuery}
                onChange={(e) => setImportQuery(e.target.value)}
                placeholder="Search for a place..."
              />
            </div>

            <div className="border border-gray-200 rounded-md max-h-64 overflow-y-auto">
              {candidates.length === 0 && (
                <div className="p-4 text-sm text-gray-500">No results yet.</div>
              )}
              {candidates.map(candidate => (
                <button
                  key={candidate.place_id}
                  type="button"
                  className={`w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 ${
                    selectedCandidate?.place_id === candidate.place_id ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => setSelectedCandidate(candidate)}
                >
                  <div className="font-medium">{candidate.name}</div>
                  <div className="text-xs text-gray-500">{candidate.formatted_address}</div>
                </button>
              ))}
            </div>

            <div className="mt-6">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-sm font-semibold">Import Status</h3>
                <button
                  type="button"
                  className="text-xs text-blue-600"
                  onClick={() => fetchImportJobs(importZoneId || undefined)}
                >
                  Refresh
                </button>
              </div>
              <div className="border border-gray-200 rounded-md max-h-40 overflow-y-auto">
                {importJobs.length === 0 && (
                  <div className="p-3 text-xs text-gray-500">No import activity yet.</div>
                )}
                {importJobs.map((job) => (
                  <div key={job.id} className="flex items-center justify-between px-3 py-2 border-b border-gray-100 text-xs">
                    <div>
                      <div className="font-medium">{job.placeId}</div>
                      <div className="text-gray-500">
                        Genre:{' '}
                        {job.genre?.name ??
                          genreNameById.get(job.genreId) ??
                          'Unknown'}
                      </div>
                      {job.errorMessage && (
                        <div className="text-red-600">{job.errorMessage}</div>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="uppercase text-[10px] text-gray-600">{job.status}</div>
                      {job.status === 'failed' && (
                        <button
                          type="button"
                          className="rounded-md border border-gray-300 px-2 py-1 text-[10px] text-gray-700"
                          onClick={() =>
                            handleRetryImport(job.placeId, job.zoneId, job.genreId)
                          }
                        >
                          Retry
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-6">
              <button
                type="button"
                className="px-4 py-2 rounded-md border border-gray-300"
                onClick={() => {
                  setShowImportModal(false);
                  resetImportForm();
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-md bg-green-600 text-white"
                onClick={handleImportPointOfInterest}
              >
                Import
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
