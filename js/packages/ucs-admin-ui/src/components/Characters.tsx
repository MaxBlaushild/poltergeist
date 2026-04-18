import { useAPI } from '@poltergeist/contexts';
import {
  Character,
  CharacterTemplate,
  PointOfInterest,
  CharacterAction,
  DialogueMessage,
  Quest,
  ZoneGenre,
} from '@poltergeist/types';
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { DialogueActionEditor } from './DialogueActionEditor.tsx';
import { DialogueMessageListEditor } from './DialogueMessageListEditor.tsx';
import {
  ShopActionEditor,
  ShopActionSavePayload,
} from './ShopActionEditor.tsx';
import { useSearchParams } from 'react-router-dom';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

type ComboOption = {
  value: string;
  label: string;
  searchText?: string;
};

type SearchableComboBoxProps = {
  label: string;
  value: string;
  options: ComboOption[];
  placeholder?: string;
  allowClear?: boolean;
  clearLabel?: string;
  onChange: (value: string) => void;
};

const normalizeDialogueEffect = (effect?: DialogueMessage['effect']) => {
  switch (effect) {
    case 'angry':
    case 'surprised':
    case 'whisper':
    case 'shout':
    case 'mysterious':
    case 'determined':
      return effect;
    default:
      return undefined;
  }
};

const SearchableComboBox: React.FC<SearchableComboBoxProps> = ({
  label,
  value,
  options,
  placeholder,
  allowClear = false,
  clearLabel = 'None',
  onChange,
}) => {
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);

  const selectedLabel =
    options.find((option) => option.value === value)?.label ?? '';

  useEffect(() => {
    if (!open) {
      setQuery(selectedLabel);
    }
  }, [selectedLabel, open]);

  const normalizedQuery = query.trim().toLowerCase();
  const filteredOptions = options.filter((option) => {
    if (!normalizedQuery) return true;
    const haystack = `${option.label} ${option.searchText ?? ''}`.toLowerCase();
    return haystack.includes(normalizedQuery);
  });

  const handleSelect = (nextValue: string, nextLabel: string) => {
    onChange(nextValue);
    setQuery(nextLabel);
    setOpen(false);
  };

  return (
    <div style={{ position: 'relative' }}>
      <label style={{ display: 'block', marginBottom: '5px' }}>{label}</label>
      <input
        type="text"
        value={query}
        placeholder={placeholder}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => {
          setTimeout(() => setOpen(false), 120);
        }}
        style={{
          width: '100%',
          padding: '8px',
          border: '1px solid #ccc',
          borderRadius: '4px',
        }}
      />
      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            zIndex: 20,
            marginTop: '4px',
            background: '#fff',
            border: '1px solid #e5e7eb',
            borderRadius: '6px',
            boxShadow: '0 8px 18px rgba(0,0,0,0.08)',
            maxHeight: '220px',
            overflowY: 'auto',
          }}
        >
          {allowClear && (
            <button
              type="button"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect('', '')}
              style={{
                width: '100%',
                textAlign: 'left',
                padding: '8px 10px',
                border: 'none',
                background: 'transparent',
                cursor: 'pointer',
                fontSize: '13px',
                color: '#6b7280',
              }}
            >
              {clearLabel}
            </button>
          )}
          {filteredOptions.length === 0 ? (
            <div
              style={{
                padding: '8px 10px',
                fontSize: '13px',
                color: '#6b7280',
              }}
            >
              No matches
            </div>
          ) : (
            filteredOptions.map((option) => (
              <button
                key={option.value}
                type="button"
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => handleSelect(option.value, option.label)}
                style={{
                  width: '100%',
                  textAlign: 'left',
                  padding: '8px 10px',
                  border: 'none',
                  background:
                    option.value === value ? '#eff6ff' : 'transparent',
                  cursor: 'pointer',
                  fontSize: '13px',
                }}
              >
                {option.label}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
};

interface CharacterLocationsMapProps {
  locations: [number, number][];
  onAddLocation: (lng: number, lat: number) => void;
  onRemoveLocation: (index: number) => void;
}

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

const normalizeInternalTags = (input: string[]) => {
  const seen = new Set<string>();
  const normalized: string[] = [];
  input.forEach((value) => {
    const tag = value.trim().toLowerCase();
    if (!tag || seen.has(tag)) return;
    seen.add(tag);
    normalized.push(tag);
  });
  return normalized;
};

const parseInternalTagsInput = (input: string) =>
  normalizeInternalTags(
    input
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean)
  );

type CharacterStoryVariantForm = {
  id?: string;
  priority: number;
  requiredStoryFlagsText: string;
  description: string;
  dialogue: DialogueMessage[];
};

const buildCharacterStoryVariantForm = (
  variant?: NonNullable<Character['storyVariants']>[number]
): CharacterStoryVariantForm => ({
  id: variant?.id,
  priority: variant?.priority ?? 0,
  requiredStoryFlagsText: (variant?.requiredStoryFlags ?? []).join(', '),
  description: variant?.description ?? '',
  dialogue: variant?.dialogue ?? [],
});

const defaultCharacterUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered character. Hidden wanderer silhouette, mysterious cloak motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

const staticStatusColor = (status?: string) => {
  const normalized = (status || '').trim().toLowerCase();
  switch (normalized) {
    case 'completed':
      return '#059669';
    case 'queued':
    case 'in_progress':
      return '#2563eb';
    case 'failed':
    case 'missing':
      return '#b91c1c';
    default:
      return '#4b5563';
  }
};

const formatDate = (value?: string | null) => {
  if (!value) return '—';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const getErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message.trim()) {
    return error.message.trim();
  }
  return fallback;
};

const defaultGenreIdFromList = (genres: ZoneGenre[]): string => {
  const fantasyGenre = genres.find(
    (genre) => genre.name?.trim().toLowerCase() === 'fantasy'
  );
  return fantasyGenre?.id ?? genres[0]?.id ?? '';
};

const formatGenreLabel = (genre?: ZoneGenre | null): string =>
  genre?.name?.trim() || 'Unknown Genre';

const CharacterLocationsMap: React.FC<CharacterLocationsMapProps> = ({
  locations,
  onAddLocation,
  onRemoveLocation,
}) => {
  const mapContainer = React.useRef<HTMLDivElement>(null);
  const map = React.useRef<mapboxgl.Map | null>(null);
  const markers = React.useRef<mapboxgl.Marker[]>([]);
  const [mapLoaded, setMapLoaded] = React.useState(false);
  const [isLocating, setIsLocating] = React.useState(false);
  const [locationError, setLocationError] = React.useState<string | null>(null);
  const [didAutoLocate, setDidAutoLocate] = React.useState(false);

  React.useEffect(() => {
    if (mapContainer.current && !map.current) {
      const initialCenter = locations.length > 0 ? locations[0] : [0, 0];
      map.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: initialCenter,
        zoom: locations.length > 0 ? 14 : 2,
        interactive: true,
      });
      map.current.on('load', () => setMapLoaded(true));
      map.current.on('click', (e) => {
        onAddLocation(e.lngLat.lng, e.lngLat.lat);
      });
    }

    return () => {
      markers.current.forEach((marker) => marker.remove());
      markers.current = [];
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, []);

  React.useEffect(() => {
    if (map.current && mapLoaded && locations.length > 0) {
      map.current.setCenter(locations[0]);
      map.current.setZoom(Math.max(map.current.getZoom(), 14));
      setDidAutoLocate(true);
    }
  }, [locations, mapLoaded]);

  React.useEffect(() => {
    if (!map.current || !mapLoaded || didAutoLocate || locations.length > 0)
      return;
    if (!navigator.geolocation) {
      setLocationError('Geolocation is not supported in this browser.');
      return;
    }
    setIsLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setIsLocating(false);
        const { latitude, longitude } = pos.coords;
        map.current?.flyTo({
          center: [longitude, latitude],
          zoom: Math.max(map.current?.getZoom() ?? 14, 16),
          essential: true,
        });
        setDidAutoLocate(true);
      },
      (err) => {
        setIsLocating(false);
        setLocationError(err.message || 'Unable to fetch location.');
        setDidAutoLocate(true);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 10000 }
    );
  }, [mapLoaded, didAutoLocate, locations.length]);

  React.useEffect(() => {
    if (!map.current || !mapLoaded) return;

    markers.current.forEach((marker) => marker.remove());
    markers.current = [];

    locations.forEach((location, index) => {
      const el = document.createElement('div');
      el.style.width = '18px';
      el.style.height = '18px';
      el.style.borderRadius = '9999px';
      el.style.background = '#2563eb';
      el.style.border = '2px solid white';
      el.style.boxShadow = '0 1px 4px rgba(0,0,0,0.3)';
      el.style.cursor = 'pointer';
      el.addEventListener('click', (event) => {
        event.stopPropagation();
        onRemoveLocation(index);
      });

      const marker = new mapboxgl.Marker(el)
        .setLngLat(location)
        .addTo(map.current!);
      markers.current.push(marker);
    });
  }, [locations, mapLoaded, onRemoveLocation]);

  return (
    <div className="relative w-full h-80 rounded-lg border border-gray-300 overflow-hidden">
      <div ref={mapContainer} className="w-full h-full" />
      <div className="absolute top-3 right-3 flex flex-col gap-2 z-10">
        <button
          type="button"
          className="bg-white border border-gray-300 rounded shadow px-2 py-1 text-sm hover:bg-gray-50"
          onClick={() => map.current?.zoomIn()}
          aria-label="Zoom in"
        >
          +
        </button>
        <button
          type="button"
          className="bg-white border border-gray-300 rounded shadow px-2 py-1 text-sm hover:bg-gray-50"
          onClick={() => map.current?.zoomOut()}
          aria-label="Zoom out"
        >
          −
        </button>
        <button
          type="button"
          className="bg-white border border-gray-300 rounded shadow px-2 py-1 text-sm hover:bg-gray-50"
          onClick={() => {
            if (!navigator.geolocation) {
              setLocationError('Geolocation is not supported in this browser.');
              return;
            }
            setIsLocating(true);
            setLocationError(null);
            navigator.geolocation.getCurrentPosition(
              (pos) => {
                setIsLocating(false);
                const { latitude, longitude } = pos.coords;
                map.current?.flyTo({
                  center: [longitude, latitude],
                  zoom: Math.max(map.current?.getZoom() ?? 14, 16),
                  essential: true,
                });
                setDidAutoLocate(true);
              },
              (err) => {
                setIsLocating(false);
                setLocationError(err.message || 'Unable to fetch location.');
                setDidAutoLocate(true);
              },
              { enableHighAccuracy: true, timeout: 10000, maximumAge: 10000 }
            );
          }}
          aria-label="Center on your location"
          disabled={isLocating}
        >
          {isLocating ? '...' : 'My Location'}
        </button>
        {locationError && (
          <div className="bg-white border border-red-200 text-red-600 text-xs rounded px-2 py-1 shadow">
            {locationError}
          </div>
        )}
      </div>
    </div>
  );
};

export const Characters = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const { apiClient } = useAPI();
  const { zoneQuestArchetypes, updateZoneQuestArchetype } =
    useQuestArchetypes();
  const [characters, setCharacters] = useState<Character[]>([]);
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [filteredCharacters, setFilteredCharacters] = useState<Character[]>([]);
  const [selectedCharacterIds, setSelectedCharacterIds] = useState<Set<string>>(
    new Set()
  );
  const [bulkDeletingCharacters, setBulkDeletingCharacters] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [genreFilter, setGenreFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateCharacter, setShowCreateCharacter] = useState(false);
  const [showGenerateCharacter, setShowGenerateCharacter] = useState(false);
  const [editingCharacter, setEditingCharacter] = useState<Character | null>(
    null
  );
  const [availablePointsOfInterest, setAvailablePointsOfInterest] = useState<
    PointOfInterest[]
  >([]);
  const [selectedZoneQuestArchetypeIds, setSelectedZoneQuestArchetypeIds] =
    useState<string[]>([]);

  // Dialogue management state
  const [selectedCharacterForDialogue, setSelectedCharacterForDialogue] =
    useState<Character | null>(null);
  const [characterActions, setCharacterActions] = useState<CharacterAction[]>(
    []
  );
  const [questById, setQuestById] = useState<Record<string, Quest | null>>({});
  const [questLookupLoading, setQuestLookupLoading] = useState(false);
  const [editingAction, setEditingAction] = useState<CharacterAction | null>(
    null
  );
  const [showDialogueEditor, setShowDialogueEditor] = useState(false);
  const [showDialogueManager, setShowDialogueManager] = useState(false);
  const [showShopEditor, setShowShopEditor] = useState(false);
  const [characterLocations, setCharacterLocations] = useState<
    [number, number][]
  >([]);
  const [savingLocations, setSavingLocations] = useState(false);
  const [savingTemplate, setSavingTemplate] = useState(false);
  const [templateSaveState, setTemplateSaveState] = useState<{
    kind: 'success' | 'error';
    message: string;
  } | null>(null);
  const [locationsError, setLocationsError] = useState<string | null>(null);
  const [generationData, setGenerationData] = useState({
    name: '',
    genreId: '',
    description: '',
  });
  const [generationError, setGenerationError] = useState<string | null>(null);
  const [generatingCharacter, setGeneratingCharacter] = useState(false);
  const [characterUndiscoveredBusy, setCharacterUndiscoveredBusy] =
    useState(false);
  const [
    characterUndiscoveredStatusLoading,
    setCharacterUndiscoveredStatusLoading,
  ] = useState(false);
  const [characterUndiscoveredError, setCharacterUndiscoveredError] = useState<
    string | null
  >(null);
  const [characterUndiscoveredMessage, setCharacterUndiscoveredMessage] =
    useState<string | null>(null);
  const [characterUndiscoveredUrl, setCharacterUndiscoveredUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png'
  );
  const [characterUndiscoveredStatus, setCharacterUndiscoveredStatus] =
    useState('unknown');
  const [characterUndiscoveredExists, setCharacterUndiscoveredExists] =
    useState(false);
  const [
    characterUndiscoveredRequestedAt,
    setCharacterUndiscoveredRequestedAt,
  ] = useState<string | null>(null);
  const [
    characterUndiscoveredLastModified,
    setCharacterUndiscoveredLastModified,
  ] = useState<string | null>(null);
  const [
    characterUndiscoveredPreviewNonce,
    setCharacterUndiscoveredPreviewNonce,
  ] = useState<number>(Date.now());
  const [characterUndiscoveredPrompt, setCharacterUndiscoveredPrompt] =
    useState(defaultCharacterUndiscoveredIconPrompt);
  const didHydrateDeepLinkedCharacterRef = React.useRef(false);
  const deepLinkedCharacterId = searchParams.get('id')?.trim() ?? '';
  const replaceDeepLinkedCharacterId = useCallback(
    (characterId?: string | null) => {
      const normalizedCharacterId = (characterId ?? '').trim();
      const currentCharacterId = searchParams.get('id')?.trim() ?? '';
      if (normalizedCharacterId === currentCharacterId) {
        return;
      }
      const next = new URLSearchParams(searchParams);
      if (normalizedCharacterId) {
        next.set('id', normalizedCharacterId);
      } else {
        next.delete('id');
      }
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams]
  );

  const pointOfInterestOptions = React.useMemo<ComboOption[]>(() => {
    return [...availablePointsOfInterest]
      .sort((a, b) => (a.name || '').localeCompare(b.name || ''))
      .map((poi) => ({
        value: poi.id,
        label: poi.name || poi.description || poi.id,
        searchText: [poi.name, poi.description, poi.id]
          .filter(Boolean)
          .join(' '),
      }));
  }, [availablePointsOfInterest]);

  const genreNameById = useMemo(() => {
    const next = new Map<string, string>();
    genres.forEach((genre) => next.set(genre.id, genre.name));
    return next;
  }, [genres]);

  const defaultGenreId = useMemo(
    () => defaultGenreIdFromList(genres),
    [genres]
  );

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    genreId: '',
    internalTagsInput: '',
    storyVariants: [] as CharacterStoryVariantForm[],
    mapIconUrl: '',
    dialogueImageUrl: '',
    thumbnailUrl: '',
    pointOfInterestId: '',
  });

  useEffect(() => {
    fetchGenres();
    fetchCharacters();
    fetchPointsOfInterest();
  }, []);

  useEffect(() => {
    if (!defaultGenreId) {
      return;
    }
    setFormData((prev) =>
      prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
    );
    setGenerationData((prev) =>
      prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
    );
  }, [defaultGenreId]);

  useEffect(() => {
    const hasPending = characters.some((character) =>
      ['queued', 'in_progress'].includes(character.imageGenerationStatus || '')
    );
    if (!hasPending) return;

    const interval = setInterval(() => {
      fetchCharacters();
    }, 5000);

    return () => clearInterval(interval);
  }, [characters]);

  const fetchPointsOfInterest = async () => {
    try {
      const response = await apiClient.get<PointOfInterest[]>(
        '/sonar/pointsOfInterest'
      );
      setAvailablePointsOfInterest(response);
    } catch (error) {
      console.error('Error fetching points of interest:', error);
    }
  };

  const fetchGenres = async () => {
    try {
      const response = await apiClient.get<ZoneGenre[]>(
        '/sonar/zone-genres?includeInactive=true'
      );
      setGenres(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching character genres:', error);
      setGenres([]);
    }
  };

  const refreshUndiscoveredCharacterIconStatus = React.useCallback(
    async (showMessage = false) => {
      try {
        setCharacterUndiscoveredStatusLoading(true);
        setCharacterUndiscoveredError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/character-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setCharacterUndiscoveredUrl(url);
        }
        setCharacterUndiscoveredStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setCharacterUndiscoveredExists(Boolean(response?.exists));
        setCharacterUndiscoveredRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setCharacterUndiscoveredLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setCharacterUndiscoveredPreviewNonce(Date.now());
        if (showMessage) {
          setCharacterUndiscoveredMessage(
            'Undiscovered character icon status refreshed.'
          );
        }
      } catch (error) {
        console.error(
          'Failed to load undiscovered character icon status',
          error
        );
        const message =
          error instanceof Error
            ? error.message
            : 'Failed to load undiscovered character icon status.';
        setCharacterUndiscoveredError(message);
      } finally {
        setCharacterUndiscoveredStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateUndiscoveredCharacterIcon =
    React.useCallback(async () => {
      const prompt = characterUndiscoveredPrompt.trim();
      if (!prompt) {
        setCharacterUndiscoveredError('Prompt is required.');
        return;
      }
      try {
        setCharacterUndiscoveredBusy(true);
        setCharacterUndiscoveredError(null);
        setCharacterUndiscoveredMessage(null);
        await apiClient.post<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/character-undiscovered',
          { prompt }
        );
        setCharacterUndiscoveredMessage(
          'Undiscovered character icon queued for generation.'
        );
        await refreshUndiscoveredCharacterIconStatus();
      } catch (error) {
        console.error('Failed to generate undiscovered character icon', error);
        const message =
          error instanceof Error
            ? error.message
            : 'Failed to generate undiscovered character icon.';
        setCharacterUndiscoveredError(message);
      } finally {
        setCharacterUndiscoveredBusy(false);
      }
    }, [
      apiClient,
      characterUndiscoveredPrompt,
      refreshUndiscoveredCharacterIconStatus,
    ]);

  const handleDeleteUndiscoveredCharacterIcon = React.useCallback(async () => {
    try {
      setCharacterUndiscoveredBusy(true);
      setCharacterUndiscoveredError(null);
      setCharacterUndiscoveredMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/character-undiscovered'
      );
      setCharacterUndiscoveredMessage('Undiscovered character icon deleted.');
      await refreshUndiscoveredCharacterIconStatus();
    } catch (error) {
      console.error('Failed to delete undiscovered character icon', error);
      const message =
        error instanceof Error
          ? error.message
          : 'Failed to delete undiscovered character icon.';
      setCharacterUndiscoveredError(message);
    } finally {
      setCharacterUndiscoveredBusy(false);
    }
  }, [apiClient, refreshUndiscoveredCharacterIconStatus]);

  useEffect(() => {
    if (searchQuery === '' && genreFilter === '') {
      setFilteredCharacters(characters);
    } else {
      const filtered = characters.filter(
        (character) => {
          const matchesSearch =
            searchQuery === '' ||
            character.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
            character.description
              ?.toLowerCase()
              .includes(searchQuery.toLowerCase()) ||
            (character.internalTags ?? []).some((tag) =>
              tag.toLowerCase().includes(searchQuery.toLowerCase())
            ) ||
            (character.genre?.name ?? genreNameById.get(character.genreId ?? ''))
              ?.toLowerCase()
              .includes(searchQuery.toLowerCase());
          const matchesGenre =
            genreFilter === '' ||
            (character.genreId ?? character.genre?.id ?? '') === genreFilter;
          return matchesSearch && matchesGenre;
        }
      );
      setFilteredCharacters(filtered);
    }
  }, [searchQuery, genreFilter, characters, genreNameById]);

  useEffect(() => {
    setSelectedCharacterIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(characters.map((character) => character.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      if (next.size === prev.size) {
        return prev;
      }
      return next;
    });
  }, [characters]);

  useEffect(() => {
    if (!editingCharacter) return;
    setSelectedZoneQuestArchetypeIds(
      zoneQuestArchetypes
        .filter(
          (zoneQuestArchetype) =>
            zoneQuestArchetype.characterId === editingCharacter.id
        )
        .map((zoneQuestArchetype) => zoneQuestArchetype.id)
    );
  }, [editingCharacter, zoneQuestArchetypes]);

  useEffect(() => {
    void refreshUndiscoveredCharacterIconStatus();
  }, [refreshUndiscoveredCharacterIconStatus]);

  useEffect(() => {
    if (
      characterUndiscoveredStatus !== 'queued' &&
      characterUndiscoveredStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshUndiscoveredCharacterIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [characterUndiscoveredStatus, refreshUndiscoveredCharacterIconStatus]);

  const fetchCharacters = async () => {
    try {
      const response = await apiClient.get<Character[]>('/sonar/characters');
      setCharacters(response);
      setFilteredCharacters(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching characters:', error);
      setLoading(false);
    }
  };

  // Dialogue management functions
  const fetchCharacterActions = async (characterId: string) => {
    try {
      const response = await apiClient.get<CharacterAction[]>(
        `/sonar/characters/${characterId}/actions`
      );
      setCharacterActions(response);
      return response;
    } catch (error) {
      console.error('Error fetching character actions:', error);
      return [];
    }
  };

  const getActionQuestId = (action: CharacterAction) => {
    const raw =
      action.metadata?.questId ?? action.metadata?.pointOfInterestGroupId;
    if (!raw) return '';
    return String(raw);
  };

  const ensureQuestLookups = async (actions: CharacterAction[]) => {
    const questIds = actions
      .filter((action) => action.actionType === 'giveQuest')
      .map(getActionQuestId)
      .filter((id): id is string => Boolean(id));

    if (questIds.length === 0) {
      return;
    }

    const uniqueQuestIds = Array.from(new Set(questIds));
    const missingQuestIds = uniqueQuestIds.filter((id) => !(id in questById));

    if (missingQuestIds.length === 0) {
      return;
    }

    setQuestLookupLoading(true);
    const results = await Promise.allSettled(
      missingQuestIds.map((id) => apiClient.get<Quest>(`/sonar/quests/${id}`))
    );

    setQuestById((prev) => {
      const next = { ...prev };
      results.forEach((result, index) => {
        const questId = missingQuestIds[index];
        if (result.status === 'fulfilled') {
          next[questId] = result.value;
        } else {
          next[questId] = null;
        }
      });
      return next;
    });
    setQuestLookupLoading(false);
  };

  const createCharacterAction = async (
    characterId: string,
    actionType: 'talk' | 'shop',
    dialogue?: DialogueMessage[],
    metadata?: any
  ) => {
    try {
      const newAction = await apiClient.post<CharacterAction>(
        '/sonar/character-actions',
        {
          characterId,
          actionType,
          dialogue: dialogue || [],
          metadata: metadata || {},
        }
      );
      setCharacterActions([...characterActions, newAction]);
      return newAction;
    } catch (error) {
      console.error('Error creating character action:', error);
      throw error;
    }
  };

  const updateCharacterAction = async (
    actionId: string,
    dialogue?: DialogueMessage[],
    metadata?: any
  ) => {
    try {
      const updates: any = {};
      if (dialogue !== undefined) {
        updates.dialogue = dialogue;
      }
      if (metadata !== undefined) {
        updates.metadata = metadata;
      }
      const updatedAction = await apiClient.put<CharacterAction>(
        `/sonar/character-actions/${actionId}`,
        updates
      );
      setCharacterActions(
        characterActions.map((a) => (a.id === actionId ? updatedAction : a))
      );
      return updatedAction;
    } catch (error) {
      console.error('Error updating character action:', error);
      throw error;
    }
  };

  const deleteCharacterAction = async (actionId: string) => {
    try {
      await apiClient.delete(`/sonar/character-actions/${actionId}`);
      setCharacterActions(characterActions.filter((a) => a.id !== actionId));
    } catch (error) {
      console.error('Error deleting character action:', error);
    }
  };

  const handleManageDialogue = async (character: Character) => {
    setSelectedCharacterForDialogue(character);
    setShowDialogueManager(true);
    const actions = await fetchCharacterActions(character.id);
    await ensureQuestLookups(actions);
  };

  const handleCreateNewAction = () => {
    setEditingAction(null);
    setShowDialogueEditor(true);
  };

  const handleEditAction = (action: CharacterAction) => {
    setEditingAction(action);
    if (action.actionType === 'shop') {
      setShowShopEditor(true);
    } else {
      setShowDialogueEditor(true);
    }
  };

  const handleSaveDialogue = async (dialogue: DialogueMessage[]) => {
    if (!selectedCharacterForDialogue) return;

    try {
      if (editingAction) {
        await updateCharacterAction(editingAction.id, dialogue);
      } else {
        await createCharacterAction(
          selectedCharacterForDialogue.id,
          'talk',
          dialogue
        );
      }
      setShowDialogueEditor(false);
      setEditingAction(null);
      const actions = await fetchCharacterActions(
        selectedCharacterForDialogue.id
      );
      await ensureQuestLookups(actions);
    } catch (error) {
      console.error('Error saving dialogue:', error);
    }
  };

  const handleSaveShop = async (payload: ShopActionSavePayload) => {
    if (!selectedCharacterForDialogue) return;

    try {
      if (editingAction) {
        await updateCharacterAction(editingAction.id, undefined, payload);
      } else {
        await createCharacterAction(
          selectedCharacterForDialogue.id,
          'shop',
          [],
          payload
        );
      }
      setShowShopEditor(false);
      setEditingAction(null);
      const actions = await fetchCharacterActions(
        selectedCharacterForDialogue.id
      );
      await ensureQuestLookups(actions);
    } catch (error) {
      console.error('Error saving shop:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      genreId: defaultGenreId,
      internalTagsInput: '',
      storyVariants: [],
      mapIconUrl: '',
      dialogueImageUrl: '',
      thumbnailUrl: '',
      pointOfInterestId: '',
    });
    setCharacterLocations([]);
    setLocationsError(null);
    setTemplateSaveState(null);
    setSelectedZoneQuestArchetypeIds([]);
  };

  const resetGenerationForm = () => {
    setGenerationData({
      name: '',
      genreId: defaultGenreId,
      description: '',
    });
    setGenerationError(null);
  };

  const buildCharacterPayload = () => {
    return {
      name: formData.name,
      description: formData.description,
      genreId: formData.genreId,
      mapIconUrl: formData.mapIconUrl,
      dialogueImageUrl: formData.dialogueImageUrl,
      thumbnailUrl: formData.thumbnailUrl,
      internalTags: parseInternalTagsInput(formData.internalTagsInput),
      storyVariants: formData.storyVariants
        .map((variant) => ({
          id: variant.id,
          priority: Number(variant.priority) || 0,
          requiredStoryFlags: parseInternalTagsInput(
            variant.requiredStoryFlagsText
          ),
          description: variant.description.trim(),
          dialogue: variant.dialogue
            .map((entry, index) => ({
              speaker: entry.speaker === 'user' ? 'user' : 'character',
              text: (entry.text ?? '').trim(),
              order: index,
              effect: normalizeDialogueEffect(entry.effect),
            }))
            .filter((entry) => entry.text.length > 0),
        }))
        .filter(
          (variant) =>
            variant.requiredStoryFlags.length > 0 ||
            variant.description.length > 0 ||
            variant.dialogue.length > 0
        ),
      pointOfInterestId: formData.pointOfInterestId || null,
    };
  };

  const applyQuestAssignments = async (
    characterId: string,
    nextZoneQuestArchetypeIds: string[]
  ) => {
    const updates: Promise<void>[] = [];
    zoneQuestArchetypes.forEach((zoneQuestArchetype) => {
      const shouldBeAssigned = nextZoneQuestArchetypeIds.includes(
        zoneQuestArchetype.id
      );
      const isAssignedToCharacter =
        zoneQuestArchetype.characterId === characterId;
      if (shouldBeAssigned && !isAssignedToCharacter) {
        updates.push(
          updateZoneQuestArchetype(zoneQuestArchetype.id, { characterId })
        );
      } else if (!shouldBeAssigned && isAssignedToCharacter) {
        updates.push(
          updateZoneQuestArchetype(zoneQuestArchetype.id, { characterId: null })
        );
      }
    });

    if (updates.length > 0) {
      await Promise.all(updates);
    }
  };

  const saveCharacterLocations = async (characterId: string) => {
    setSavingLocations(true);
    setLocationsError(null);
    try {
      await apiClient.put(`/sonar/characters/${characterId}/locations`, {
        locations: characterLocations.map(([lng, lat]) => ({
          latitude: lat,
          longitude: lng,
        })),
      });
      await fetchCharacters();
    } catch (error) {
      console.error('Error saving character locations:', error);
      setLocationsError('Failed to save character locations.');
    } finally {
      setSavingLocations(false);
    }
  };

  const handleCreateCharacter = async () => {
    try {
      const payload = buildCharacterPayload();
      const newCharacter = await apiClient.post<Character>(
        '/sonar/characters',
        payload
      );
      setCharacters([...characters, newCharacter]);
      await applyQuestAssignments(
        newCharacter.id,
        selectedZoneQuestArchetypeIds
      );
      if (characterLocations.length > 0) {
        await saveCharacterLocations(newCharacter.id);
      }
      setShowCreateCharacter(false);
      resetForm();
    } catch (error) {
      console.error('Error creating character:', error);
    }
  };

  const handleUpdateCharacter = async () => {
    if (!editingCharacter) return;

    try {
      const payload = buildCharacterPayload();
      const updatedCharacter = await apiClient.put<Character>(
        `/sonar/characters/${editingCharacter.id}`,
        payload
      );
      setCharacters(
        characters.map((c) =>
          c.id === editingCharacter.id ? updatedCharacter : c
        )
      );
      await applyQuestAssignments(
        editingCharacter.id,
        selectedZoneQuestArchetypeIds
      );
      await saveCharacterLocations(editingCharacter.id);
      setEditingCharacter(null);
      resetForm();
    } catch (error) {
      console.error('Error updating character:', error);
    }
  };

  const handleGenerateCharacter = async () => {
    const name = generationData.name.trim();
    const description = generationData.description.trim();
    if (!name) {
      setGenerationError('Name is required.');
      return;
    }

    try {
      setGeneratingCharacter(true);
      setGenerationError(null);
      const newCharacter = await apiClient.post<Character>(
        '/sonar/characters/generate',
        {
          name,
          genreId: generationData.genreId,
          description,
        }
      );
      setCharacters([...characters, newCharacter]);
      setShowGenerateCharacter(false);
      resetGenerationForm();
    } catch (error) {
      console.error('Error generating character:', error);
      setGenerationError(
        getErrorMessage(error, 'Failed to generate character.')
      );
    } finally {
      setGeneratingCharacter(false);
    }
  };

  const handleRegenerateCharacterImage = async (character: Character) => {
    try {
      const updated = await apiClient.post<Character>(
        `/sonar/characters/${character.id}/regenerate`,
        {}
      );
      setCharacters(
        characters.map((c) => (c.id === character.id ? updated : c))
      );
    } catch (error) {
      console.error('Error regenerating character image:', error);
      alert('Error regenerating character image.');
    }
  };

  const handleDeleteCharacter = async (character: Character) => {
    try {
      await apiClient.delete(`/sonar/characters/${character.id}`);
      setCharacters(characters.filter((c) => c.id !== character.id));
      setSelectedCharacterIds((prev) => {
        const next = new Set(prev);
        next.delete(character.id);
        return next;
      });
    } catch (error) {
      console.error('Error deleting character:', error);
    }
  };

  const handleSaveCharacterTemplate = async () => {
    try {
      setSavingTemplate(true);
      setTemplateSaveState(null);
      const template = await apiClient.post<CharacterTemplate>(
        '/sonar/character-templates',
        buildCharacterPayload()
      );
      setTemplateSaveState({
        kind: 'success',
        message: `Saved "${template.name || 'character'}" as a reusable character template.`,
      });
    } catch (error) {
      console.error('Error saving character template:', error);
      setTemplateSaveState({
        kind: 'error',
        message: getErrorMessage(error, 'Failed to save character template.'),
      });
    } finally {
      setSavingTemplate(false);
    }
  };

  const allFilteredSelected =
    filteredCharacters.length > 0 &&
    filteredCharacters.every((character) =>
      selectedCharacterIds.has(character.id)
    );

  const toggleSelectAllFiltered = () => {
    setSelectedCharacterIds((prev) => {
      const next = new Set(prev);
      if (allFilteredSelected) {
        filteredCharacters.forEach((character) => next.delete(character.id));
      } else {
        filteredCharacters.forEach((character) => next.add(character.id));
      }
      return next;
    });
  };

  const toggleCharacterSelection = (characterID: string) => {
    setSelectedCharacterIds((prev) => {
      const next = new Set(prev);
      if (next.has(characterID)) {
        next.delete(characterID);
      } else {
        next.add(characterID);
      }
      return next;
    });
  };

  const handleBulkDeleteCharacters = async () => {
    if (bulkDeletingCharacters || selectedCharacterIds.size === 0) return;

    const selectedIDs = Array.from(selectedCharacterIds);
    const selectedNames = characters
      .filter((character) => selectedCharacterIds.has(character.id))
      .map((character) => character.name || character.id);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIDs.length === 1
        ? `Delete 1 selected character (${preview})? This cannot be undone.`
        : `Delete ${selectedIDs.length} selected characters${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;
    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingCharacters(true);
    try {
      await apiClient.post('/sonar/characters/bulk-delete', {
        ids: selectedIDs,
      });
      const deleted = new Set(selectedIDs);
      setCharacters((prev) =>
        prev.filter((character) => !deleted.has(character.id))
      );
      setSelectedCharacterIds(new Set());
      if (editingCharacter && deleted.has(editingCharacter.id)) {
        setEditingCharacter(null);
        resetForm();
      }
      if (
        selectedCharacterForDialogue &&
        deleted.has(selectedCharacterForDialogue.id)
      ) {
        setSelectedCharacterForDialogue(null);
        setShowDialogueManager(false);
        setShowDialogueEditor(false);
        setShowShopEditor(false);
        setEditingAction(null);
        setCharacterActions([]);
      }
    } catch (error) {
      console.error('Error bulk deleting characters:', error);
      alert('Failed to bulk delete characters.');
    } finally {
      setBulkDeletingCharacters(false);
    }
  };

  const handleEditCharacter = useCallback(
    (character: Character) => {
      setEditingCharacter(character);
      setFormData({
        name: character.name,
        description: character.description,
        genreId: character.genreId ?? character.genre?.id ?? defaultGenreId,
        internalTagsInput: (character.internalTags ?? []).join(', '),
        storyVariants: (character.storyVariants ?? []).map((variant) =>
          buildCharacterStoryVariantForm(variant)
        ),
        mapIconUrl: character.mapIconUrl,
        dialogueImageUrl: character.dialogueImageUrl,
        thumbnailUrl: character.thumbnailUrl ?? '',
        pointOfInterestId: character.pointOfInterestId ?? '',
      });
      const locations =
        character.locations?.map(
          (loc) => [loc.longitude, loc.latitude] as [number, number]
        ) ?? [];
      setCharacterLocations(locations);
      setLocationsError(null);
      setSelectedZoneQuestArchetypeIds(
        zoneQuestArchetypes
          .filter(
            (zoneQuestArchetype) =>
              zoneQuestArchetype.characterId === character.id
          )
          .map((zoneQuestArchetype) => zoneQuestArchetype.id)
      );
    },
    [defaultGenreId, zoneQuestArchetypes]
  );

  useEffect(() => {
    if (didHydrateDeepLinkedCharacterRef.current) {
      return;
    }
    if (!deepLinkedCharacterId) {
      didHydrateDeepLinkedCharacterRef.current = true;
      return;
    }
    if (editingCharacter?.id === deepLinkedCharacterId) {
      didHydrateDeepLinkedCharacterRef.current = true;
      return;
    }
    const matchingCharacter = characters.find(
      (character) => character.id === deepLinkedCharacterId
    );
    if (matchingCharacter) {
      handleEditCharacter(matchingCharacter);
      return;
    }
    if (!characters.length) {
      return;
    }
    void apiClient
      .get<Character>(`/sonar/characters/${deepLinkedCharacterId}`)
      .then((character) => {
        if (!character) {
          didHydrateDeepLinkedCharacterRef.current = true;
          return;
        }
        setCharacters((prev) => {
          if (prev.some((entry) => entry.id === character.id)) {
            return prev;
          }
          return [character, ...prev];
        });
        handleEditCharacter(character);
      })
      .catch((error) => {
        console.error('Failed to deep link character', error);
        didHydrateDeepLinkedCharacterRef.current = true;
      });
  }, [
    apiClient,
    characters,
    deepLinkedCharacterId,
    editingCharacter,
    handleEditCharacter,
  ]);

  useEffect(() => {
    if (!didHydrateDeepLinkedCharacterRef.current) {
      return;
    }
    replaceDeepLinkedCharacterId(editingCharacter?.id ?? null);
  }, [editingCharacter, replaceDeepLinkedCharacterId]);

  const addCharacterLocation = (lng: number, lat: number) => {
    setCharacterLocations((prev) => [...prev, [lng, lat]]);
  };

  const removeCharacterLocation = (index: number) => {
    setCharacterLocations((prev) => prev.filter((_, i) => i !== index));
  };

  const formatGenerationStatus = (status?: string) => {
    switch (status) {
      case 'queued':
        return 'Queued';
      case 'in_progress':
        return 'Generating';
      case 'complete':
        return 'Complete';
      case 'failed':
        return 'Failed';
      case 'none':
        return 'Not requested';
      default:
        return 'Unknown';
    }
  };

  if (loading) {
    return <div className="m-10">Loading characters...</div>;
  }

  return (
    <div className="m-10">
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
          flexWrap: 'wrap',
          marginBottom: '16px',
        }}
      >
        <h1 className="text-2xl font-bold" style={{ margin: 0 }}>
          Characters
        </h1>
        <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-md"
            onClick={() => {
              resetForm();
              setShowCreateCharacter(true);
            }}
          >
            Create Character
          </button>
          <button
            className="bg-green-600 text-white px-4 py-2 rounded-md"
            onClick={() => {
              resetGenerationForm();
              setShowGenerateCharacter(true);
            }}
          >
            Generate Character
          </button>
        </div>
      </div>

      <div
        style={{
          marginBottom: '20px',
          padding: '16px',
          border: '1px solid #d1d5db',
          borderRadius: '8px',
          backgroundColor: '#ffffff',
          boxShadow: '0 1px 2px rgba(0,0,0,0.06)',
        }}
      >
        <div
          style={{
            display: 'flex',
            gap: '10px',
            flexWrap: 'wrap',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '10px',
          }}
        >
          <h2 style={{ margin: 0, fontSize: '18px', fontWeight: 600 }}>
            Undiscovered Character Icon
          </h2>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <button
              type="button"
              className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={() => void refreshUndiscoveredCharacterIconStatus(true)}
              disabled={characterUndiscoveredStatusLoading}
            >
              {characterUndiscoveredStatusLoading
                ? 'Refreshing…'
                : 'Refresh Status'}
            </button>
            <button
              type="button"
              className="bg-indigo-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleGenerateUndiscoveredCharacterIcon}
              disabled={
                characterUndiscoveredBusy || characterUndiscoveredStatusLoading
              }
            >
              {characterUndiscoveredBusy ? 'Working…' : 'Generate Icon'}
            </button>
            <button
              type="button"
              className="bg-red-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleDeleteUndiscoveredCharacterIcon}
              disabled={
                characterUndiscoveredBusy || characterUndiscoveredStatusLoading
              }
            >
              {characterUndiscoveredBusy ? 'Working…' : 'Delete Icon'}
            </button>
          </div>
        </div>
        <div style={{ marginBottom: '6px' }}>
          <span
            style={{
              display: 'inline-flex',
              color: '#ffffff',
              fontSize: '12px',
              padding: '2px 8px',
              borderRadius: '9999px',
              backgroundColor: staticStatusColor(characterUndiscoveredStatus),
            }}
          >
            {characterUndiscoveredStatus || 'unknown'}
          </span>
        </div>
        <div
          style={{ fontSize: '12px', color: '#4b5563', wordBreak: 'break-all' }}
        >
          URL: {characterUndiscoveredUrl}
        </div>
        <div style={{ fontSize: '12px', color: '#4b5563', marginTop: '4px' }}>
          Requested: {formatDate(characterUndiscoveredRequestedAt)}
          {' · '}
          Last updated: {formatDate(characterUndiscoveredLastModified)}
        </div>
        <div style={{ marginTop: '10px' }}>
          <label style={{ display: 'block', marginBottom: '5px' }}>
            Generation Prompt
          </label>
          <textarea
            value={characterUndiscoveredPrompt}
            onChange={(event) =>
              setCharacterUndiscoveredPrompt(event.target.value)
            }
            placeholder="Prompt used to generate the undiscovered character icon."
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #d1d5db',
              borderRadius: '6px',
              minHeight: '88px',
            }}
          />
        </div>
        {characterUndiscoveredExists ? (
          <div style={{ marginTop: '12px' }}>
            <img
              src={`${characterUndiscoveredUrl}?v=${characterUndiscoveredPreviewNonce}`}
              alt="Undiscovered character icon preview"
              style={{
                width: '96px',
                height: '96px',
                borderRadius: '6px',
                border: '1px solid #d1d5db',
                objectFit: 'cover',
                backgroundColor: '#f9fafb',
              }}
            />
          </div>
        ) : (
          <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '8px' }}>
            No icon currently found at this URL.
          </div>
        )}
        {characterUndiscoveredMessage ? (
          <div style={{ fontSize: '14px', color: '#047857', marginTop: '8px' }}>
            {characterUndiscoveredMessage}
          </div>
        ) : null}
        {characterUndiscoveredError ? (
          <div style={{ fontSize: '14px', color: '#b91c1c', marginTop: '8px' }}>
            {characterUndiscoveredError}
          </div>
        ) : null}
      </div>

      {/* Search */}
      <div className="mb-4">
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'minmax(0, 1fr) 240px',
            gap: '12px',
          }}
        >
          <input
            type="text"
            placeholder="Search characters..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full p-2 border rounded-md"
          />
          <select
            value={genreFilter}
            onChange={(e) => setGenreFilter(e.target.value)}
            className="w-full p-2 border rounded-md"
          >
            <option value="">All genres</option>
            {genres.map((genre) => (
              <option key={`character-genre-filter-${genre.id}`} value={genre.id}>
                {formatGenreLabel(genre)}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '12px',
          marginBottom: '12px',
          flexWrap: 'wrap',
        }}
      >
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontSize: '14px',
          }}
        >
          <input
            type="checkbox"
            checked={allFilteredSelected}
            onChange={toggleSelectAllFiltered}
            disabled={filteredCharacters.length === 0 || bulkDeletingCharacters}
          />
          Select all filtered ({filteredCharacters.length})
        </label>
        <button
          className="bg-red-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
          onClick={handleBulkDeleteCharacters}
          disabled={selectedCharacterIds.size === 0 || bulkDeletingCharacters}
        >
          {bulkDeletingCharacters
            ? `Deleting ${selectedCharacterIds.size}...`
            : `Delete Selected (${selectedCharacterIds.size})`}
        </button>
      </div>

      {/* Characters Grid */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
          gap: '20px',
          padding: '20px',
        }}
      >
        {filteredCharacters.map((character) => (
          <div
            key={character.id}
            style={{
              padding: '20px',
              border: '1px solid #ccc',
              borderRadius: '8px',
              backgroundColor: '#fff',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            }}
          >
            <div style={{ marginBottom: '10px' }}>
              <label
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '8px',
                  fontSize: '13px',
                }}
              >
                <input
                  type="checkbox"
                  checked={selectedCharacterIds.has(character.id)}
                  onChange={() => toggleCharacterSelection(character.id)}
                />
                Select
              </label>
            </div>
            <h2
              style={{
                margin: '0 0 15px 0',
                color: '#333',
              }}
            >
              {character.name}
            </h2>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Description: {character.description || 'No description'}
            </p>
            <p style={{ margin: '5px 0', color: '#666' }}>
              Genre:{' '}
              {formatGenreLabel(
                character.genre ??
                  genres.find((genre) => genre.id === character.genreId)
              )}
            </p>
            <p style={{ margin: '5px 0', color: '#666' }}>
              Internal Tags:{' '}
              {(character.internalTags ?? []).join(', ') || 'None'}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Dialogue Image URL: {character.dialogueImageUrl || '—'}
            </p>
            <p style={{ margin: '5px 0', color: '#666' }}>
              Image Status:{' '}
              {formatGenerationStatus(character.imageGenerationStatus)}
            </p>
            {character.imageGenerationStatus === 'failed' &&
              character.imageGenerationError && (
                <p
                  style={{
                    margin: '5px 0',
                    color: '#b91c1c',
                    fontSize: '12px',
                  }}
                >
                  Error: {character.imageGenerationError}
                </p>
              )}
            {character.dialogueImageUrl && (
              <img
                src={character.dialogueImageUrl}
                alt={`${character.name} dialogue`}
                style={{ maxWidth: '100%', maxHeight: 120, borderRadius: 4 }}
              />
            )}

            <div style={{ marginTop: '15px' }}>
              <button
                onClick={() => handleEditCharacter(character)}
                className="bg-blue-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Edit
              </button>
              <button
                onClick={() => handleManageDialogue(character)}
                className="bg-green-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Manage Dialogue
              </button>
              <button
                onClick={() => handleRegenerateCharacterImage(character)}
                className="bg-yellow-500 text-white px-4 py-2 rounded-md mr-2"
                disabled={['queued', 'in_progress'].includes(
                  character.imageGenerationStatus || ''
                )}
              >
                Regenerate Image
              </button>
              <button
                onClick={() => handleDeleteCharacter(character)}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create/Edit Character Modal */}
      {(showCreateCharacter || editingCharacter) && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '600px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <h2>{editingCharacter ? 'Edit Character' : 'Create Character'}</h2>
            {templateSaveState ? (
              <div
                style={{
                  marginTop: '10px',
                  marginBottom: '15px',
                  padding: '10px 12px',
                  borderRadius: '6px',
                  border:
                    templateSaveState.kind === 'success'
                      ? '1px solid #a7f3d0'
                      : '1px solid #fecaca',
                  backgroundColor:
                    templateSaveState.kind === 'success'
                      ? '#ecfdf5'
                      : '#fef2f2',
                  color:
                    templateSaveState.kind === 'success'
                      ? '#047857'
                      : '#b91c1c',
                  fontSize: '14px',
                }}
              >
                {templateSaveState.message}
              </div>
            ) : null}

            {/* Character Fields */}
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Name:
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Description:
              </label>
              <textarea
                value={formData.description}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                  minHeight: '60px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Genre:
              </label>
              <select
                value={formData.genreId}
                onChange={(e) =>
                  setFormData({ ...formData, genreId: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              >
                {genres.map((genre) => (
                  <option key={`character-form-genre-${genre.id}`} value={genre.id}>
                    {formatGenreLabel(genre)}
                  </option>
                ))}
              </select>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Internal Tags:
              </label>
              <input
                type="text"
                value={formData.internalTagsInput}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    internalTagsInput: e.target.value,
                  })
                }
                placeholder="merchant, starter_quest, blacksmith"
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <div
                style={{ marginTop: '6px', color: '#666', fontSize: '12px' }}
              >
                Comma-separated metadata tags used to link characters with quest
                templates.
              </div>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '8px' }}>
                Story Variants
              </label>
              <div style={{ fontSize: '12px', color: '#666', marginBottom: 8 }}>
                Optional conditional overrides for this NPC&apos;s description
                and dialogue, keyed off story flags.
              </div>
              {formData.storyVariants.length === 0 ? (
                <div
                  style={{ fontSize: '12px', color: '#999', marginBottom: 8 }}
                >
                  No story variants yet.
                </div>
              ) : (
                formData.storyVariants.map((variant, index) => (
                  <div
                    key={variant.id ?? `story-variant-${index}`}
                    style={{
                      border: '1px solid #d1d5db',
                      borderRadius: '6px',
                      padding: '12px',
                      marginBottom: '10px',
                      backgroundColor: '#f9fafb',
                    }}
                  >
                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: '120px 1fr',
                        gap: '10px',
                        marginBottom: '10px',
                      }}
                    >
                      <div>
                        <label
                          style={{ display: 'block', marginBottom: '5px' }}
                        >
                          Priority
                        </label>
                        <input
                          type="number"
                          value={variant.priority}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              storyVariants: prev.storyVariants.map(
                                (entry, storyIndex) =>
                                  storyIndex === index
                                    ? {
                                        ...entry,
                                        priority:
                                          parseInt(e.target.value, 10) || 0,
                                      }
                                    : entry
                              ),
                            }))
                          }
                          style={{
                            width: '100%',
                            padding: '8px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{ display: 'block', marginBottom: '5px' }}
                        >
                          Required Story Flags
                        </label>
                        <input
                          type="text"
                          value={variant.requiredStoryFlagsText}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              storyVariants: prev.storyVariants.map(
                                (entry, storyIndex) =>
                                  storyIndex === index
                                    ? {
                                        ...entry,
                                        requiredStoryFlagsText: e.target.value,
                                      }
                                    : entry
                              ),
                            }))
                          }
                          placeholder="chapter_2_started, warden_warned"
                          style={{
                            width: '100%',
                            padding: '8px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                    </div>
                    <div style={{ marginBottom: '10px' }}>
                      <label style={{ display: 'block', marginBottom: '5px' }}>
                        Description Override
                      </label>
                      <textarea
                        value={variant.description}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            storyVariants: prev.storyVariants.map(
                              (entry, storyIndex) =>
                                storyIndex === index
                                  ? {
                                      ...entry,
                                      description: e.target.value,
                                    }
                                  : entry
                            ),
                          }))
                        }
                        style={{
                          width: '100%',
                          padding: '8px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                          minHeight: '70px',
                        }}
                      />
                    </div>
                    <div style={{ marginBottom: '10px' }}>
                      <DialogueMessageListEditor
                        label="Dialogue Override"
                        helperText="Story variants can also use line effects like Angry."
                        value={variant.dialogue}
                        onChange={(dialogue) =>
                          setFormData((prev) => ({
                            ...prev,
                            storyVariants: prev.storyVariants.map(
                              (entry, storyIndex) =>
                                storyIndex === index
                                  ? {
                                      ...entry,
                                      dialogue,
                                    }
                                  : entry
                            ),
                          }))
                        }
                      />
                    </div>
                    <button
                      type="button"
                      className="bg-red-100 text-red-700 px-3 py-1 rounded-md"
                      onClick={() =>
                        setFormData((prev) => ({
                          ...prev,
                          storyVariants: prev.storyVariants.filter(
                            (_, storyIndex) => storyIndex !== index
                          ),
                        }))
                      }
                    >
                      Remove Variant
                    </button>
                  </div>
                ))
              )}
              <button
                type="button"
                className="bg-gray-100 text-gray-800 px-3 py-2 rounded-md"
                onClick={() =>
                  setFormData((prev) => ({
                    ...prev,
                    storyVariants: [
                      ...prev.storyVariants,
                      buildCharacterStoryVariantForm(),
                    ],
                  }))
                }
              >
                Add Story Variant
              </button>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Map Icon URL:
              </label>
              <input
                type="text"
                value={formData.mapIconUrl}
                onChange={(e) =>
                  setFormData({ ...formData, mapIconUrl: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Thumbnail URL:
              </label>
              <input
                type="text"
                value={formData.thumbnailUrl}
                onChange={(e) =>
                  setFormData({ ...formData, thumbnailUrl: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Dialogue Image URL:
              </label>
              <input
                type="text"
                value={formData.dialogueImageUrl}
                onChange={(e) =>
                  setFormData({ ...formData, dialogueImageUrl: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Character Locations
              </label>
              <div
                style={{
                  marginBottom: '10px',
                  color: '#666',
                  fontSize: '12px',
                }}
              >
                Click on the map to add a pin. Click an existing pin to remove
                it.
              </div>
              <CharacterLocationsMap
                locations={characterLocations}
                onAddLocation={addCharacterLocation}
                onRemoveLocation={removeCharacterLocation}
              />
              <div style={{ marginTop: '10px' }}>
                <div
                  style={{
                    fontSize: '12px',
                    color: '#666',
                    marginBottom: '6px',
                  }}
                >
                  Saved locations ({characterLocations.length})
                </div>
                {characterLocations.length === 0 ? (
                  <div style={{ fontSize: '12px', color: '#999' }}>
                    No locations yet.
                  </div>
                ) : (
                  <div
                    style={{
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '6px',
                    }}
                  >
                    {characterLocations.map(([lng, lat], index) => (
                      <div
                        key={`${lng}-${lat}-${index}`}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          padding: '6px 8px',
                          border: '1px solid #e5e7eb',
                          borderRadius: '4px',
                          fontSize: '12px',
                        }}
                      >
                        <span>
                          {lat.toFixed(6)}, {lng.toFixed(6)}
                        </span>
                        <button
                          type="button"
                          onClick={() => removeCharacterLocation(index)}
                          style={{
                            border: 'none',
                            background: 'transparent',
                            color: '#c00',
                            cursor: 'pointer',
                          }}
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
              {locationsError && (
                <div
                  style={{ marginTop: '8px', color: '#c00', fontSize: '12px' }}
                >
                  {locationsError}
                </div>
              )}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'flex-end',
                  marginTop: '10px',
                }}
              >
                <button
                  type="button"
                  onClick={() => {
                    if (!editingCharacter) {
                      setLocationsError(
                        'Save the character first to store locations.'
                      );
                      return;
                    }
                    saveCharacterLocations(editingCharacter.id);
                  }}
                  disabled={savingLocations}
                  style={{
                    padding: '8px 12px',
                    backgroundColor: '#2563eb',
                    color: '#fff',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: savingLocations ? 'default' : 'pointer',
                    opacity: savingLocations ? 0.7 : 1,
                  }}
                >
                  {savingLocations ? 'Saving...' : 'Save Locations'}
                </button>
              </div>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <SearchableComboBox
                label="Point of Interest (optional):"
                value={formData.pointOfInterestId}
                options={pointOfInterestOptions}
                placeholder="Search points of interest..."
                allowClear
                clearLabel="None"
                onChange={(value) =>
                  setFormData({ ...formData, pointOfInterestId: value })
                }
              />
            </div>

            {/* Quest Associations */}
            <div
              style={{
                marginBottom: '15px',
                padding: '15px',
                border: '1px solid #eee',
                borderRadius: '4px',
              }}
            >
              <h3 style={{ margin: '0 0 15px 0' }}>Quest Associations</h3>
              {zoneQuestArchetypes.length === 0 ? (
                <div style={{ color: '#999', fontStyle: 'italic' }}>
                  No zone quest archetypes available.
                </div>
              ) : (
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '10px',
                    maxHeight: '240px',
                    overflow: 'auto',
                  }}
                >
                  {zoneQuestArchetypes.map((zoneQuestArchetype) => {
                    const isAssigned = selectedZoneQuestArchetypeIds.includes(
                      zoneQuestArchetype.id
                    );
                    const assignedCharacter =
                      zoneQuestArchetype.character?.name ||
                      characters.find(
                        (c) => c.id === zoneQuestArchetype.characterId
                      )?.name ||
                      'Unassigned';
                    return (
                      <label
                        key={zoneQuestArchetype.id}
                        style={{
                          display: 'flex',
                          alignItems: 'flex-start',
                          gap: '10px',
                          padding: '10px',
                          border: '1px solid #ddd',
                          borderRadius: '6px',
                          backgroundColor: isAssigned ? '#f0f8ff' : '#fff',
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={isAssigned}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedZoneQuestArchetypeIds([
                                ...selectedZoneQuestArchetypeIds,
                                zoneQuestArchetype.id,
                              ]);
                            } else {
                              setSelectedZoneQuestArchetypeIds(
                                selectedZoneQuestArchetypeIds.filter(
                                  (id) => id !== zoneQuestArchetype.id
                                )
                              );
                            }
                          }}
                        />
                        <div style={{ flex: 1 }}>
                          <div style={{ fontWeight: 600 }}>
                            {zoneQuestArchetype.questArchetype?.name ||
                              zoneQuestArchetype.questArchetypeId}
                          </div>
                          <div style={{ color: '#666', fontSize: '13px' }}>
                            Zone:{' '}
                            {zoneQuestArchetype.zone?.name ||
                              zoneQuestArchetype.zoneId}
                          </div>
                          <div style={{ color: '#666', fontSize: '13px' }}>
                            Number of Quests:{' '}
                            {zoneQuestArchetype.numberOfQuests}
                          </div>
                          <div style={{ color: '#999', fontSize: '12px' }}>
                            Current quest giver: {assignedCharacter}
                          </div>
                        </div>
                      </label>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Modal Buttons */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'flex-end',
                gap: '10px',
              }}
            >
              <button
                onClick={handleSaveCharacterTemplate}
                className="bg-white text-slate-900 px-4 py-2 rounded-md border border-slate-300"
                disabled={savingTemplate}
              >
                {savingTemplate ? 'Saving Template...' : 'Save As Template'}
              </button>
              <button
                onClick={() => {
                  setShowCreateCharacter(false);
                  setEditingCharacter(null);
                  resetForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
              <button
                onClick={
                  editingCharacter
                    ? handleUpdateCharacter
                    : handleCreateCharacter
                }
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingCharacter ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Generate Character Modal */}
      {showGenerateCharacter && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '500px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <h2>Generate Character</h2>
            <p
              style={{
                marginTop: '8px',
                marginBottom: '16px',
                color: '#4b5563',
              }}
            >
              Queue a new character and their image generation job.
            </p>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Name *:
              </label>
              <input
                type="text"
                value={generationData.name}
                onChange={(e) =>
                  setGenerationData({ ...generationData, name: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Description:
              </label>
              <textarea
                value={generationData.description}
                onChange={(e) =>
                  setGenerationData({
                    ...generationData,
                    description: e.target.value,
                  })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                  minHeight: '80px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Genre:
              </label>
              <select
                value={generationData.genreId}
                onChange={(e) =>
                  setGenerationData({
                    ...generationData,
                    genreId: e.target.value,
                  })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              >
                {genres.map((genre) => (
                  <option
                    key={`character-generation-genre-${genre.id}`}
                    value={genre.id}
                  >
                    {formatGenreLabel(genre)}
                  </option>
                ))}
              </select>
            </div>

            {generationError ? (
              <div
                style={{
                  marginBottom: '15px',
                  padding: '10px 12px',
                  borderRadius: '6px',
                  backgroundColor: '#fef2f2',
                  border: '1px solid #fecaca',
                  color: '#b91c1c',
                  fontSize: '14px',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {generationError}
              </div>
            ) : null}

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={handleGenerateCharacter}
                className="bg-green-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
                disabled={generatingCharacter}
              >
                {generatingCharacter ? 'Generating…' : 'Generate'}
              </button>
              <button
                onClick={() => {
                  setShowGenerateCharacter(false);
                  resetGenerationForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                disabled={generatingCharacter}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dialogue Manager Modal */}
      {showDialogueManager && selectedCharacterForDialogue && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '800px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <h2 style={{ margin: '0 0 20px 0' }}>
              Manage Dialogue - {selectedCharacterForDialogue.name}
            </h2>

            {/* Character Actions List */}
            <div style={{ marginBottom: '20px' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: '15px',
                }}
              >
                <h3 style={{ margin: 0 }}>Existing Actions</h3>
                <div style={{ display: 'flex', gap: '10px' }}>
                  <button
                    onClick={() => {
                      setEditingAction(null);
                      setShowDialogueEditor(true);
                    }}
                    className="bg-blue-500 text-white px-4 py-2 rounded-md"
                  >
                    Create Talk Action
                  </button>
                  <button
                    onClick={() => {
                      setEditingAction(null);
                      setShowShopEditor(true);
                    }}
                    className="bg-green-500 text-white px-4 py-2 rounded-md"
                  >
                    Create Shop Action
                  </button>
                </div>
              </div>

              {characterActions.length === 0 ? (
                <div
                  style={{
                    padding: '40px',
                    textAlign: 'center',
                    color: '#999',
                    fontStyle: 'italic',
                    border: '1px dashed #ccc',
                    borderRadius: '8px',
                  }}
                >
                  No actions yet. Create one to get started.
                </div>
              ) : (
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '10px',
                  }}
                >
                  {characterActions.map((action) => {
                    const questId = getActionQuestId(action);
                    const hasQuestLookup = questId
                      ? Object.prototype.hasOwnProperty.call(questById, questId)
                      : false;
                    const questRecord =
                      questId && hasQuestLookup
                        ? questById[questId]
                        : undefined;
                    const editDisabled = action.actionType === 'giveQuest';

                    return (
                      <div
                        key={action.id}
                        style={{
                          padding: '15px',
                          border: '1px solid #ccc',
                          borderRadius: '8px',
                          backgroundColor: '#f9f9f9',
                        }}
                      >
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                          }}
                        >
                          <div style={{ flex: 1 }}>
                            <div
                              style={{
                                fontWeight: 'bold',
                                marginBottom: '5px',
                              }}
                            >
                              Type: {action.actionType}
                            </div>
                            <div style={{ color: '#666', fontSize: '14px' }}>
                              {action.actionType === 'talk' ? (
                                action.dialogue.length > 0 ? (
                                  <>
                                    Preview:{' '}
                                    {action.dialogue[0].text.substring(0, 100)}
                                    {action.dialogue[0].text.length > 100
                                      ? '...'
                                      : ''}
                                  </>
                                ) : (
                                  'No dialogue messages'
                                )
                              ) : action.actionType === 'shop' ? (
                                action.metadata?.shopMode === 'tags' ? (
                                  <>
                                    Shop by tags (
                                    {action.metadata?.shopItemTags?.length ?? 0}{' '}
                                    tag
                                    {(action.metadata?.shopItemTags?.length ??
                                      0) !== 1
                                      ? 's'
                                      : ''}
                                    )
                                  </>
                                ) : action.metadata?.inventory ? (
                                  <>
                                    Shop with {action.metadata.inventory.length}{' '}
                                    item
                                    {action.metadata.inventory.length !== 1
                                      ? 's'
                                      : ''}
                                  </>
                                ) : (
                                  'Shop with no items'
                                )
                              ) : action.actionType === 'giveQuest' ? (
                                questId ? (
                                  hasQuestLookup ? (
                                    questRecord ? (
                                      <>Gives quest: {questRecord.name}</>
                                    ) : (
                                      <>Gives missing quest ({questId})</>
                                    )
                                  ) : questLookupLoading ? (
                                    <>Loading quest {questId}...</>
                                  ) : (
                                    <>Gives quest: {questId}</>
                                  )
                                ) : (
                                  'Give quest action (missing quest id)'
                                )
                              ) : (
                                'Unknown action type'
                              )}
                            </div>
                            <div
                              style={{
                                color: '#999',
                                fontSize: '12px',
                                marginTop: '5px',
                              }}
                            >
                              {action.actionType === 'talk' ? (
                                <>
                                  {action.dialogue.length} message
                                  {action.dialogue.length !== 1 ? 's' : ''}
                                </>
                              ) : action.actionType === 'shop' ? (
                                action.metadata?.shopMode === 'tags' ? (
                                  <>
                                    Tags:{' '}
                                    {(action.metadata?.shopItemTags ?? []).join(
                                      ', '
                                    ) || 'None'}
                                  </>
                                ) : (
                                  <>
                                    {action.metadata?.inventory?.length || 0}{' '}
                                    item
                                    {action.metadata?.inventory?.length !== 1
                                      ? 's'
                                      : ''}
                                  </>
                                )
                              ) : action.actionType === 'giveQuest' ? (
                                questId ? (
                                  <>Quest ID: {questId}</>
                                ) : (
                                  <>Missing quest metadata</>
                                )
                              ) : null}
                            </div>
                          </div>
                          <div style={{ display: 'flex', gap: '10px' }}>
                            <button
                              onClick={() => handleEditAction(action)}
                              className={
                                editDisabled
                                  ? 'bg-gray-300 text-gray-600 px-3 py-1 rounded-md cursor-not-allowed'
                                  : 'bg-blue-500 text-white px-3 py-1 rounded-md'
                              }
                              disabled={editDisabled}
                              title={
                                editDisabled
                                  ? 'GiveQuest actions are managed by quest assignments.'
                                  : 'Edit action'
                              }
                            >
                              Edit
                            </button>
                            <button
                              onClick={() => deleteCharacterAction(action.id)}
                              className="bg-red-500 text-white px-3 py-1 rounded-md"
                            >
                              Delete
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Modal Buttons */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'flex-end',
                gap: '10px',
              }}
            >
              <button
                onClick={() => {
                  setShowDialogueManager(false);
                  setSelectedCharacterForDialogue(null);
                  setCharacterActions([]);
                  setQuestById({});
                  setQuestLookupLoading(false);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dialogue Editor Modal */}
      {showDialogueEditor && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 2000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '900px',
              maxHeight: '90vh',
              overflow: 'hidden',
            }}
          >
            <h2 style={{ margin: '0 0 20px 0' }}>
              {editingAction
                ? 'Edit Dialogue Action'
                : 'Create New Dialogue Action'}
            </h2>
            <DialogueActionEditor
              action={editingAction}
              onSave={handleSaveDialogue}
              onCancel={() => {
                setShowDialogueEditor(false);
                setEditingAction(null);
              }}
            />
          </div>
        </div>
      )}

      {/* Shop Editor Modal */}
      {showShopEditor && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 2000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '900px',
              maxHeight: '90vh',
              overflow: 'hidden',
            }}
          >
            <h2 style={{ margin: '0 0 20px 0' }}>
              {editingAction ? 'Edit Shop Action' : 'Create New Shop Action'}
            </h2>
            <ShopActionEditor
              action={editingAction}
              onSave={handleSaveShop}
              onCancel={() => {
                setShowShopEditor(false);
                setEditingAction(null);
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
};
