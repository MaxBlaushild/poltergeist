import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

type ZoneSeedPointOfInterestDraft = {
  draftId: string;
  placeId: string;
  name: string;
  address?: string;
  types?: string[];
  latitude?: number;
  longitude?: number;
  rating?: number;
  userRatingCount?: number;
  editorialSummary?: string;
};

type ZoneSeedCharacterDraft = {
  draftId: string;
  name: string;
  description: string;
  placeId: string;
  latitude?: number;
  longitude?: number;
  shopItemTags?: string[];
};

type ZoneSeedDraft = {
  fantasyName?: string;
  zoneDescription?: string;
  pointsOfInterest?: ZoneSeedPointOfInterestDraft[];
  characters?: ZoneSeedCharacterDraft[];
};

type SeedMode = 'manual' | 'auto';

type ZoneSeedCounts = {
  placeCount: number;
  monsterCount: number;
  bossEncounterCount: number;
  raidEncounterCount: number;
  inputEncounterCount: number;
  optionEncounterCount: number;
  treasureChestCount: number;
  healingFountainCount: number;
};

type SeedCountField = keyof ZoneSeedCounts;

type SeedCountInputMap = Record<SeedCountField, string>;

type ZoneSeedAutoAudit = {
  zoneAreaSquareFeet?: number;
  zoneAreaAcres?: number;
  recommendedCounts?: ZoneSeedCounts;
  finalCounts?: ZoneSeedCounts;
  warnings?: string[];
};

type ZoneSeedJob = {
  id: string;
  zoneId: string;
  status: string;
  seedMode?: SeedMode;
  errorMessage?: string;
  placeCount: number;
  characterCount: number;
  questCount: number;
  mainQuestCount: number;
  monsterCount: number;
  bossEncounterCount: number;
  raidEncounterCount: number;
  inputEncounterCount: number;
  optionEncounterCount: number;
  treasureChestCount?: number;
  healingFountainCount?: number;
  requiredPlaceTags?: string[];
  shopkeeperItemTags?: string[];
  createdAt?: string;
  updatedAt?: string;
  autoSeedAudit?: ZoneSeedAutoAudit;
  draft?: ZoneSeedDraft;
};

type ZoneSeedDraftPayload = Partial<ZoneSeedCounts> & {
  seedMode?: SeedMode;
  requiredPlaceTags: string[];
  shopkeeperItemTags: string[];
};

type BulkQueueZoneSeedJobsResponse = {
  queuedCount: number;
  requestedZoneCount: number;
  jobs: ZoneSeedJob[];
};

const defaultSeedCountInputs: SeedCountInputMap = {
  placeCount: '0',
  monsterCount: '0',
  bossEncounterCount: '0',
  raidEncounterCount: '0',
  inputEncounterCount: '0',
  optionEncounterCount: '0',
  treasureChestCount: '0',
  healingFountainCount: '0',
};

const seedCountFields: Array<{
  key: SeedCountField;
  label: string;
  description: string;
}> = [
  {
    key: 'placeCount',
    label: 'Places',
    description: 'POIs and coupled standalone challenges',
  },
  {
    key: 'monsterCount',
    label: 'Monster encounters',
    description: 'Random scalable encounters',
  },
  {
    key: 'bossEncounterCount',
    label: 'Boss encounters',
    description: 'Random scalable boss encounters',
  },
  {
    key: 'raidEncounterCount',
    label: 'Raid encounters',
    description: 'Random 5-player raid encounters',
  },
  {
    key: 'inputEncounterCount',
    label: 'Input scenarios',
    description: 'Random scalable input scenarios',
  },
  {
    key: 'optionEncounterCount',
    label: 'Option scenarios',
    description: 'Random scalable option scenarios',
  },
  {
    key: 'treasureChestCount',
    label: 'Treasure chests',
    description: 'Random scalable reward chests',
  },
  {
    key: 'healingFountainCount',
    label: 'Healing fountains',
    description: 'Random healing fountain placements',
  },
];

const sortBoundaryPoints = (points: [number, number][]): [number, number][] => {
  if (points.length < 3) return points.slice();

  const centroid = points.reduce(
    (acc, point) => [
      acc[0] + point[0] / points.length,
      acc[1] + point[1] / points.length,
    ],
    [0, 0]
  );

  const sorted = points.slice().sort((a, b) => {
    const angleA = Math.atan2(a[1] - centroid[1], a[0] - centroid[0]);
    const angleB = Math.atan2(b[1] - centroid[1], b[0] - centroid[0]);
    return angleA - angleB;
  });

  const first = sorted[0];
  const last = sorted[sorted.length - 1];
  if (first[0] !== last[0] || first[1] !== last[1]) {
    sorted.push([first[0], first[1]]);
  }

  return sorted;
};

const getZoneRing = (zone: Zone): [number, number][] => {
  if (zone.points?.length) {
    return sortBoundaryPoints(
      zone.points.map(
        (point) => [point.longitude, point.latitude] as [number, number]
      )
    );
  }

  if (zone.boundaryCoords?.length) {
    return sortBoundaryPoints(
      zone.boundaryCoords.map(
        (coord) => [coord.longitude, coord.latitude] as [number, number]
      )
    );
  }

  return [];
};

type BulkZoneSelectionMapProps = {
  zones: Zone[];
  matchingZoneIds: Set<string>;
  selectedZoneIds: Set<string>;
  onToggleZone: (zoneId: string) => void;
};

const BulkZoneSelectionMap: React.FC<BulkZoneSelectionMapProps> = ({
  zones,
  matchingZoneIds,
  selectedZoneIds,
  onToggleZone,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const searchMarkerRef = useRef<mapboxgl.Marker | null>(null);
  const fitBoundsRef = useRef(false);
  const [mapLoaded, setMapLoaded] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [showSearchSuggestions, setShowSearchSuggestions] = useState(false);
  const [searchStatus, setSearchStatus] = useState<string | null>(null);
  const [searchCandidates, setSearchCandidates] = useState<
    Array<{
      id: string;
      center: [number, number];
      placeName: string;
    }>
  >([]);

  const focusMapOnLocation = useCallback(
    (lngLat: [number, number], zoom: number) => {
      if (!mapRef.current) {
        return;
      }

      searchMarkerRef.current?.remove();
      searchMarkerRef.current = new mapboxgl.Marker({ color: '#2563EB' })
        .setLngLat(lngLat)
        .addTo(mapRef.current);

      mapRef.current.flyTo({
        center: lngLat,
        zoom: Math.max(mapRef.current.getZoom(), zoom),
        essential: true,
      });
    },
    []
  );

  const handleSearchLocation = useCallback(async () => {
    const query = searchQuery.trim();
    if (!query) {
      setSearchStatus('Enter a city, neighborhood, address, or landmark.');
      return;
    }

    if (!mapboxgl.accessToken) {
      setSearchStatus(
        'Location search is unavailable because the Mapbox token is missing.'
      );
      return;
    }

    setIsSearching(true);
    setSearchStatus(null);
    setShowSearchSuggestions(false);

    try {
      const response = await fetch(
        `https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(query)}.json?access_token=${encodeURIComponent(mapboxgl.accessToken)}&limit=6&types=country,region,postcode,district,place,locality,neighborhood,address,poi`
      );

      if (!response.ok) {
        throw new Error('Search request failed.');
      }

      const data = (await response.json()) as {
        features?: Array<{
          id?: string;
          center?: [number, number];
          place_name?: string;
        }>;
      };

      const candidates = (data.features ?? [])
        .filter((feature) => feature.center && feature.center.length >= 2)
        .map((feature, index) => ({
          id: feature.id || `${feature.place_name || 'candidate'}-${index}`,
          center: [feature.center![0], feature.center![1]] as [number, number],
          placeName: feature.place_name || 'Unknown location',
        }));

      if (candidates.length === 0) {
        setSearchCandidates([]);
        setSearchStatus(`No location match found for "${query}".`);
        return;
      }

      setSearchCandidates(candidates);
      setShowSearchSuggestions(true);
      setSearchStatus('Select a result below to move the map.');
    } catch (error) {
      console.error('Error searching for location:', error);
      setSearchCandidates([]);
      setSearchStatus('Unable to search for that location right now.');
    } finally {
      setIsSearching(false);
    }
  }, [searchQuery]);

  useEffect(() => {
    if (mapContainer.current && !mapRef.current) {
      mapRef.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [-87.6298, 41.8781],
        zoom: 10,
        interactive: true,
      });

      mapRef.current.addControl(new mapboxgl.NavigationControl(), 'top-right');
      mapRef.current.on('load', () => setMapLoaded(true));
    }

    return () => {
      searchMarkerRef.current?.remove();
      searchMarkerRef.current = null;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!mapRef.current || !mapLoaded) {
      return;
    }

    const features = zones
      .map((zone) => {
        const ring = getZoneRing(zone);
        if (ring.length < 4) {
          return null;
        }

        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: [ring],
          },
          properties: {
            id: zone.id,
            name: zone.name,
            selected: selectedZoneIds.has(zone.id),
            matching: matchingZoneIds.has(zone.id),
          },
        };
      })
      .filter(Boolean);

    const geojson = {
      type: 'FeatureCollection' as const,
      features: features as Array<GeoJSON.Feature<GeoJSON.Polygon>>,
    };

    const existingSource = mapRef.current.getSource('bulk-zone-selector') as
      | mapboxgl.GeoJSONSource
      | undefined;

    if (existingSource) {
      existingSource.setData(geojson);
    } else {
      mapRef.current.addSource('bulk-zone-selector', {
        type: 'geojson',
        data: geojson,
      });

      mapRef.current.addLayer({
        id: 'bulk-zone-selector-fill',
        type: 'fill',
        source: 'bulk-zone-selector',
        paint: {
          'fill-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#7c3aed',
            ['boolean', ['get', 'matching'], false],
            '#2563eb',
            '#94a3b8',
          ],
          'fill-opacity': [
            'case',
            ['boolean', ['get', 'selected'], false],
            0.45,
            ['boolean', ['get', 'matching'], false],
            0.18,
            0.08,
          ],
        },
      });

      mapRef.current.addLayer({
        id: 'bulk-zone-selector-outline',
        type: 'line',
        source: 'bulk-zone-selector',
        paint: {
          'line-color': [
            'case',
            ['boolean', ['get', 'selected'], false],
            '#5b21b6',
            ['boolean', ['get', 'matching'], false],
            '#1d4ed8',
            '#64748b',
          ],
          'line-width': [
            'case',
            ['boolean', ['get', 'selected'], false],
            3,
            1.5,
          ],
        },
      });

      mapRef.current.on('click', 'bulk-zone-selector-fill', (event) => {
        const feature = event.features?.[0];
        const zoneId = feature?.properties?.id;
        if (typeof zoneId === 'string' && zoneId) {
          onToggleZone(zoneId);
        }
      });

      mapRef.current.on('mouseenter', 'bulk-zone-selector-fill', () => {
        mapRef.current?.getCanvas().style.setProperty('cursor', 'pointer');
      });
      mapRef.current.on('mouseleave', 'bulk-zone-selector-fill', () => {
        mapRef.current?.getCanvas().style.setProperty('cursor', '');
      });
    }

    if (!fitBoundsRef.current && features.length > 0) {
      const bounds = new mapboxgl.LngLatBounds();
      features.forEach((feature) => {
        feature.geometry.coordinates[0].forEach((coord) => {
          bounds.extend(coord as [number, number]);
        });
      });
      mapRef.current.fitBounds(bounds, { padding: 36, maxZoom: 12 });
      fitBoundsRef.current = true;
    }
  }, [zones, selectedZoneIds, matchingZoneIds, mapLoaded, onToggleZone]);

  return (
    <div className="mt-4 rounded-lg border border-gray-200 bg-white p-3">
      <div className="mb-2">
        <div className="text-sm font-semibold text-gray-900">
          Visual zone selection
        </div>
        <div className="text-xs text-gray-500">
          Click zone polygons to add or remove them from the bulk queue.
        </div>
      </div>
      <div className="mb-3 flex flex-wrap gap-2">
        <div className="relative min-w-[240px] flex-1">
          <input
            className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
            value={searchQuery}
            onChange={(event) => {
              const value = event.target.value;
              setSearchQuery(value);
              setSearchStatus(null);
              if (value.trim() === '') {
                setSearchCandidates([]);
                setShowSearchSuggestions(false);
              }
            }}
            onFocus={() => {
              if (searchCandidates.length > 0) {
                setShowSearchSuggestions(true);
              }
            }}
            onBlur={() => {
              setTimeout(() => setShowSearchSuggestions(false), 120);
            }}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault();
                void handleSearchLocation();
              }
            }}
            placeholder="Search for a city, neighborhood, address, or landmark"
          />
          {showSearchSuggestions && searchCandidates.length > 0 && (
            <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
              {searchCandidates.map((candidate) => (
                <button
                  key={candidate.id}
                  type="button"
                  onClick={() => {
                    setSearchQuery(candidate.placeName);
                    setShowSearchSuggestions(false);
                    focusMapOnLocation(candidate.center, 13);
                    setSearchStatus(`Moved map to ${candidate.placeName}.`);
                  }}
                  className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                >
                  {candidate.placeName}
                </button>
              ))}
            </div>
          )}
        </div>
        <button
          type="button"
          onClick={() => void handleSearchLocation()}
          className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          disabled={isSearching}
        >
          {isSearching ? 'Searching...' : 'Search'}
        </button>
      </div>
      <div
        ref={mapContainer}
        className="h-[320px] w-full overflow-hidden rounded border border-gray-200"
      />
      {searchStatus && (
        <p className="mt-2 text-xs text-gray-500">{searchStatus}</p>
      )}
    </div>
  );
};

const statusBadgeClass = (status: string) => {
  switch (status) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'awaiting_approval':
      return 'bg-indigo-600';
    case 'approved':
      return 'bg-indigo-700';
    case 'applying':
      return 'bg-amber-700';
    case 'applied':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    default:
      return 'bg-gray-600';
  }
};

const zoneSeedJobStateOptions = [
  'queued',
  'in_progress',
  'awaiting_approval',
  'approved',
  'applying',
  'applied',
  'failed',
] as const;

const canDeleteZoneSeedJob = (job: ZoneSeedJob) =>
  job.status !== 'in_progress' && job.status !== 'applying';

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const getJobFinalCounts = (job: ZoneSeedJob): ZoneSeedCounts => ({
  placeCount: job.placeCount ?? 0,
  monsterCount: job.monsterCount ?? 0,
  bossEncounterCount: job.bossEncounterCount ?? 0,
  raidEncounterCount: job.raidEncounterCount ?? 0,
  inputEncounterCount: job.inputEncounterCount ?? 0,
  optionEncounterCount: job.optionEncounterCount ?? 0,
  treasureChestCount: job.treasureChestCount ?? 0,
  healingFountainCount: job.healingFountainCount ?? 0,
});

const formatZoneSeedCounts = (counts: ZoneSeedCounts) =>
  `${counts.placeCount} POIs/challenges, ${counts.monsterCount} monster encounters, ${counts.bossEncounterCount} boss encounters, ${counts.raidEncounterCount} raid encounters, ${counts.inputEncounterCount} input scenarios, ${counts.optionEncounterCount} option scenarios, ${counts.treasureChestCount} treasure chests, ${counts.healingFountainCount} healing fountains`;

const autoSeedEarthRadiusMeters = 6378137;
const autoSeedSquareFeetPerSquareMeter = 10.763910416709722;
const autoSeedSquareFeetPerAcre = 43560;

const toRadians = (value: number) => (value * Math.PI) / 180;

const ringAreaSquareMeters = (ring: [number, number][]) => {
  if (ring.length < 3) return 0;

  const closedRing =
    ring[0][0] === ring[ring.length - 1][0] &&
    ring[0][1] === ring[ring.length - 1][1]
      ? ring
      : [...ring, ring[0]];

  let area = 0;
  for (let i = 0; i < closedRing.length; i += 1) {
    let lowIndex = i;
    let middleIndex = i + 1;
    let highIndex = i + 2;

    if (i === closedRing.length - 3) {
      lowIndex = closedRing.length - 3;
      middleIndex = closedRing.length - 2;
      highIndex = 0;
    } else if (i === closedRing.length - 2) {
      lowIndex = closedRing.length - 2;
      middleIndex = 0;
      highIndex = 0;
    } else if (i === closedRing.length - 1) {
      lowIndex = 0;
      middleIndex = 0;
      highIndex = 1;
    }

    area +=
      (toRadians(closedRing[highIndex][0]) -
        toRadians(closedRing[lowIndex][0])) *
      Math.sin(toRadians(closedRing[middleIndex][1]));
  }

  return Math.abs(
    (-area * autoSeedEarthRadiusMeters * autoSeedEarthRadiusMeters) / 2
  );
};

const getZoneAreaSquareFeet = (zone?: Zone) => {
  if (!zone) return null;
  const ring = getZoneRing(zone);
  if (ring.length < 3) return null;
  const squareMeters = ringAreaSquareMeters(ring);
  if (!Number.isFinite(squareMeters) || squareMeters <= 0) return null;
  return squareMeters * autoSeedSquareFeetPerSquareMeter;
};

const inferAutoCount = (areaAcres: number, multiplier: number) => {
  const count = Math.round(Math.log1p(Math.max(areaAcres, 0)) * multiplier);
  return Math.max(1, count);
};

const inferAutoSeedCounts = (
  areaSquareFeet?: number | null,
  requiredPlaceTags: string[] = []
): ZoneSeedCounts | null => {
  if (!areaSquareFeet || areaSquareFeet <= 0) {
    return null;
  }

  const areaAcres = areaSquareFeet / autoSeedSquareFeetPerAcre;
  const placeCount = Math.max(
    inferAutoCount(areaAcres, 2.75),
    requiredPlaceTags.length
  );

  return {
    placeCount,
    monsterCount: inferAutoCount(areaAcres, 1.9),
    bossEncounterCount: inferAutoCount(areaAcres, 0.85),
    raidEncounterCount: inferAutoCount(areaAcres, 0.55),
    inputEncounterCount: inferAutoCount(areaAcres, 1.1),
    optionEncounterCount: inferAutoCount(areaAcres, 1.1),
    treasureChestCount: inferAutoCount(areaAcres, 1.35),
    healingFountainCount: inferAutoCount(areaAcres, 0.75),
  };
};

const formatSquareFeet = (value?: number) => {
  if (!value || value <= 0) return 'n/a';
  return `${Math.round(value).toLocaleString()} sq ft`;
};

const formatAcres = (value?: number) => {
  if (!value || value <= 0) return 'n/a';
  return `${value.toFixed(value >= 10 ? 1 : 2)} acres`;
};

const parseCountInput = (value: string) => {
  const parsed = Number.parseInt(value, 10);
  return Number.isNaN(parsed) ? null : parsed;
};

const buildSeedDraftPayload = (params: {
  seedMode: SeedMode;
  manualCountInputs: SeedCountInputMap;
  autoOverrideInputs: Partial<SeedCountInputMap>;
  autoRecommendation: ZoneSeedCounts | null;
  requiredPlaceTags: string[];
  shopkeeperItemTags: string[];
}): { payload?: ZoneSeedDraftPayload; error?: string } => {
  if (params.seedMode === 'manual') {
    const payload: ZoneSeedDraftPayload = {
      seedMode: 'manual',
      requiredPlaceTags: params.requiredPlaceTags,
      shopkeeperItemTags: params.shopkeeperItemTags,
    };

    for (const field of seedCountFields) {
      const parsed = parseCountInput(params.manualCountInputs[field.key]);
      if (parsed === null) {
        return { error: 'Counts must be integers.' };
      }
      payload[field.key] = parsed;
    }

    return { payload };
  }

  if (!params.autoRecommendation) {
    return { error: 'Auto mode needs a zone with valid boundary geometry.' };
  }

  const payload: ZoneSeedDraftPayload = {
    seedMode: 'auto',
    requiredPlaceTags: params.requiredPlaceTags,
    shopkeeperItemTags: params.shopkeeperItemTags,
  };

  for (const field of seedCountFields) {
    const override = params.autoOverrideInputs[field.key];
    if (override == null) {
      continue;
    }
    const parsed = parseCountInput(override);
    if (parsed === null) {
      return { error: 'Counts must be integers.' };
    }
    payload[field.key] = parsed;
  }

  return { payload };
};

export const ZoneSeedJobs = () => {
  const { apiClient } = useAPI();
  const { zones, refreshZones } = useZoneContext();
  const [draftZoneId, setDraftZoneId] = useState<string>('');
  const [jobFilterZoneId, setJobFilterZoneId] = useState<string>('');
  const [jobFilterStatuses, setJobFilterStatuses] = useState<string[]>([]);
  const [jobs, setJobs] = useState<ZoneSeedJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [creatingDraft, setCreatingDraft] = useState(false);
  const [creatingBulkDrafts, setCreatingBulkDrafts] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [retryingId, setRetryingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [bulkDeletingJobs, setBulkDeletingJobs] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [seedMode, setSeedMode] = useState<SeedMode>('manual');
  const [manualCountInputs, setManualCountInputs] = useState<SeedCountInputMap>(
    defaultSeedCountInputs
  );
  const [autoOverrideInputs, setAutoOverrideInputs] = useState<
    Partial<SeedCountInputMap>
  >({});
  const [requiredPlaceTags, setRequiredPlaceTags] = useState<string[]>([]);
  const [requiredTagQuery, setRequiredTagQuery] = useState('');
  const [showRequiredTagSuggestions, setShowRequiredTagSuggestions] =
    useState(false);
  const [shopkeeperItemTags, setShopkeeperItemTags] = useState<string[]>([]);
  const [shopkeeperTagQuery, setShopkeeperTagQuery] = useState('');
  const [bulkZoneQuery, setBulkZoneQuery] = useState('');
  const [bulkSelectedZoneIds, setBulkSelectedZoneIds] = useState<string[]>([]);
  const [selectedJobIds, setSelectedJobIds] = useState<Set<string>>(new Set());

  const knownPlaceTags = useMemo(
    () => [
      'cafe',
      'coffee_shop',
      'bakery',
      'restaurant',
      'bar',
      'ice_cream_shop',
      'dessert',
      'park',
      'garden',
      'playground',
      'trail',
      'hiking_area',
      'natural_feature',
      'beach',
      'plaza',
      'square',
      'bridge',
      'museum',
      'art_gallery',
      'gallery',
      'library',
      'book_store',
      'movie_theater',
      'theater',
      'music_venue',
      'stadium',
      'sports_complex',
      'amusement_park',
      'zoo',
      'aquarium',
      'market',
      'shopping_mall',
      'store',
      'clothing_store',
      'florist',
    ],
    []
  );
  const [draftZoneQuery, setDraftZoneQuery] = useState('');
  const [showDraftZoneSuggestions, setShowDraftZoneSuggestions] =
    useState(false);
  const [filterZoneQuery, setFilterZoneQuery] = useState('');
  const [showFilterZoneSuggestions, setShowFilterZoneSuggestions] =
    useState(false);
  const sortedZones = useMemo(
    () => [...zones].sort((left, right) => left.name.localeCompare(right.name)),
    [zones]
  );
  const zoneNameById = useMemo(() => {
    const entries = sortedZones.map((zone) => [zone.id, zone.name] as const);
    return new Map(entries);
  }, [sortedZones]);

  const selectedZone = useMemo<Zone | undefined>(() => {
    return sortedZones.find((zone) => zone.id === draftZoneId);
  }, [sortedZones, draftZoneId]);

  const autoAreaSquareFeet = useMemo(
    () => getZoneAreaSquareFeet(selectedZone),
    [selectedZone]
  );

  const autoAreaAcres = useMemo(
    () =>
      autoAreaSquareFeet
        ? autoAreaSquareFeet / autoSeedSquareFeetPerAcre
        : null,
    [autoAreaSquareFeet]
  );

  const autoRecommendation = useMemo(
    () => inferAutoSeedCounts(autoAreaSquareFeet, requiredPlaceTags),
    [autoAreaSquareFeet, requiredPlaceTags]
  );

  const autoDisplayedCountInputs = useMemo(() => {
    return seedCountFields.reduce((acc, field) => {
      const override = autoOverrideInputs[field.key];
      if (override != null) {
        acc[field.key] = override;
        return acc;
      }

      const recommendedValue = autoRecommendation?.[field.key];
      acc[field.key] =
        typeof recommendedValue === 'number' ? String(recommendedValue) : '';
      return acc;
    }, {} as SeedCountInputMap);
  }, [autoOverrideInputs, autoRecommendation]);

  const autoFinalPreviewCounts = useMemo<ZoneSeedCounts | null>(() => {
    if (!autoRecommendation) {
      return null;
    }

    const nextCounts = { ...autoRecommendation };
    for (const field of seedCountFields) {
      const override = autoOverrideInputs[field.key];
      if (override == null) {
        continue;
      }
      const parsed = parseCountInput(override);
      if (parsed == null) {
        return null;
      }
      nextCounts[field.key] = parsed;
    }

    return nextCounts;
  }, [autoOverrideInputs, autoRecommendation]);

  const draftZoneSuggestions = useMemo(() => {
    const query = draftZoneQuery.toLowerCase();
    return sortedZones.filter((zone) =>
      zone.name.toLowerCase().includes(query)
    );
  }, [sortedZones, draftZoneQuery]);

  const filterZoneSuggestions = useMemo(() => {
    const query = filterZoneQuery.toLowerCase();
    return sortedZones.filter((zone) =>
      zone.name.toLowerCase().includes(query)
    );
  }, [sortedZones, filterZoneQuery]);

  const bulkMatchingZones = useMemo(() => {
    const query = bulkZoneQuery.trim().toLowerCase();
    if (!query) {
      return sortedZones;
    }
    return sortedZones.filter((zone) =>
      zone.name.toLowerCase().includes(query)
    );
  }, [sortedZones, bulkZoneQuery]);

  const bulkTargetZones = useMemo(() => {
    const selectedZoneIds = new Set(bulkSelectedZoneIds);
    return sortedZones.filter((zone) => selectedZoneIds.has(zone.id));
  }, [sortedZones, bulkSelectedZoneIds]);

  const bulkMatchingZoneIds = useMemo(
    () => new Set(bulkMatchingZones.map((zone) => zone.id)),
    [bulkMatchingZones]
  );

  const bulkSelectedZoneIdSet = useMemo(
    () => new Set(bulkSelectedZoneIds),
    [bulkSelectedZoneIds]
  );

  useEffect(() => {
    if (sortedZones.length === 0) {
      refreshZones();
      return;
    }
    if (!draftZoneId && sortedZones.length > 0) {
      setDraftZoneId(sortedZones[0].id);
    }
  }, [sortedZones, draftZoneId, refreshZones]);

  const fetchJobs = useCallback(
    async (zoneId?: string, statuses: string[] = []) => {
      setLoadingJobs(true);
      setError(null);
      try {
        const query: Record<string, string | number> = { limit: 25 };
        if (zoneId) {
          query.zoneId = zoneId;
        }
        if (statuses.length > 0) {
          query.statuses = statuses.join(',');
        }
        const response = await apiClient.get<ZoneSeedJob[]>(
          '/sonar/admin/zone-seed-jobs',
          query
        );
        setJobs(response);
      } catch (err) {
        console.error('Failed to load zone seed jobs', err);
        setError('Failed to load zone seed jobs.');
      } finally {
        setLoadingJobs(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    fetchJobs(jobFilterZoneId || undefined, jobFilterStatuses);
  }, [fetchJobs, jobFilterZoneId, jobFilterStatuses]);

  useEffect(() => {
    if (selectedZone?.name) {
      setDraftZoneQuery(selectedZone.name);
    }
  }, [selectedZone]);

  useEffect(() => {
    setAutoOverrideInputs({});
  }, [draftZoneId]);

  useEffect(() => {
    const validZoneIds = new Set(sortedZones.map((zone) => zone.id));
    setBulkSelectedZoneIds((prev) =>
      prev.filter((zoneId) => validZoneIds.has(zoneId))
    );
  }, [sortedZones]);

  useEffect(() => {
    const visibleDeletableJobIds = new Set(
      jobs.filter((job) => canDeleteZoneSeedJob(job)).map((job) => job.id)
    );
    setSelectedJobIds((prev) => {
      const next = new Set<string>();
      prev.forEach((id) => {
        if (visibleDeletableJobIds.has(id)) {
          next.add(id);
        }
      });
      return next;
    });
  }, [jobs]);

  const visibleDeletableJobs = useMemo(
    () => jobs.filter((job) => canDeleteZoneSeedJob(job)),
    [jobs]
  );

  const updateManualCountInput = useCallback(
    (field: SeedCountField, value: string) => {
      setManualCountInputs((prev) => ({
        ...prev,
        [field]: value,
      }));
    },
    []
  );

  const updateAutoOverrideInput = useCallback(
    (field: SeedCountField, value: string) => {
      setAutoOverrideInputs((prev) => {
        const recommendedValue = autoRecommendation?.[field];
        if (recommendedValue != null && value === String(recommendedValue)) {
          const next = { ...prev };
          delete next[field];
          return next;
        }

        return {
          ...prev,
          [field]: value,
        };
      });
    },
    [autoRecommendation]
  );

  const allVisibleDeletableJobsSelected =
    visibleDeletableJobs.length > 0 &&
    visibleDeletableJobs.every((job) => selectedJobIds.has(job.id));

  const toggleJobFilterStatus = useCallback((status: string) => {
    setJobFilterStatuses((prev) =>
      prev.includes(status)
        ? prev.filter((existing) => existing !== status)
        : [...prev, status]
    );
  }, []);

  const clearJobSelection = useCallback(() => {
    setSelectedJobIds(new Set());
  }, []);

  const handleToggleJobSelection = useCallback((jobId: string) => {
    setSelectedJobIds((prev) => {
      const next = new Set(prev);
      if (next.has(jobId)) {
        next.delete(jobId);
      } else {
        next.add(jobId);
      }
      return next;
    });
  }, []);

  const handleSelectVisibleJobs = useCallback(() => {
    setSelectedJobIds((prev) => {
      const next = new Set(prev);
      visibleDeletableJobs.forEach((job) => next.add(job.id));
      return next;
    });
  }, [visibleDeletableJobs]);

  const handleBulkDeleteJobs = useCallback(async () => {
    if (bulkDeletingJobs || selectedJobIds.size === 0 || deletingId !== null) {
      return;
    }

    const selectedIds = Array.from(selectedJobIds);
    const selectedLabels = jobs
      .filter((job) => selectedJobIds.has(job.id))
      .map(
        (job) =>
          `${zoneNameById.get(job.zoneId) || job.zoneId} · ${job.id.slice(0, 8)}`
      );
    const preview = selectedLabels.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedLabels.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected zone seed job (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected zone seed jobs${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingJobs(true);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post('/sonar/admin/zone-seed-jobs/bulk-delete', {
        ids: selectedIds,
      });
      const deletedIds = new Set(selectedIds);
      setJobs((prev) => prev.filter((job) => !deletedIds.has(job.id)));
      setSelectedJobIds(new Set());
      setSuccess(
        selectedIds.length === 1
          ? 'Draft job deleted.'
          : `Deleted ${selectedIds.length} draft jobs.`
      );
    } catch (err) {
      console.error('Failed to bulk delete zone seed jobs', err);
      setError('Failed to delete selected draft jobs.');
    } finally {
      setBulkDeletingJobs(false);
    }
  }, [
    apiClient,
    bulkDeletingJobs,
    deletingId,
    jobs,
    selectedJobIds,
    zoneNameById,
  ]);

  const handleCreateDraft = async () => {
    if (!draftZoneId) {
      setError('Please select a zone.');
      return;
    }
    const { payload, error: payloadError } = buildSeedDraftPayload({
      seedMode,
      manualCountInputs,
      autoOverrideInputs,
      autoRecommendation,
      requiredPlaceTags,
      shopkeeperItemTags,
    });
    if (!payload) {
      setError(payloadError ?? 'Counts must be integers.');
      return;
    }
    setCreatingDraft(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<ZoneSeedJob>(
        `/sonar/admin/zones/${draftZoneId}/seed-draft`,
        payload
      );
      setJobs((prev) => [created, ...prev]);
      const warningCount = created.autoSeedAudit?.warnings?.length ?? 0;
      setSuccess(
        warningCount > 0
          ? `Draft queued successfully with ${warningCount} auto-mode note${warningCount === 1 ? '' : 's'}.`
          : 'Draft queued successfully.'
      );
    } catch (err) {
      console.error('Failed to queue draft', err);
      setError('Failed to queue zone seed draft.');
    } finally {
      setCreatingDraft(false);
    }
  };

  const handleBulkCreateDrafts = async () => {
    if (bulkTargetZones.length === 0) {
      setError('Select at least one zone for bulk queueing.');
      return;
    }
    if (seedMode !== 'manual') {
      setError('Auto mode is only available for single-zone drafts.');
      return;
    }

    const { payload, error: payloadError } = buildSeedDraftPayload({
      seedMode: 'manual',
      manualCountInputs,
      autoOverrideInputs: {},
      autoRecommendation: null,
      requiredPlaceTags,
      shopkeeperItemTags,
    });
    if (!payload) {
      setError(payloadError ?? 'Counts must be integers.');
      return;
    }

    setCreatingBulkDrafts(true);
    setError(null);
    setSuccess(null);
    try {
      const response = await apiClient.post<BulkQueueZoneSeedJobsResponse>(
        '/sonar/admin/zone-seed-jobs/bulk-queue',
        {
          zoneIds: bulkTargetZones.map((zone) => zone.id),
          ...payload,
        }
      );
      setJobs((prev) => [...response.jobs, ...prev]);
      if (response.queuedCount === response.requestedZoneCount) {
        setSuccess(`Queued ${response.queuedCount} zone seed draft jobs.`);
      } else {
        setSuccess(
          `Queued ${response.queuedCount} zone seed draft jobs from ${response.requestedZoneCount} requested zones.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk queue drafts', err);
      setError('Failed to bulk queue zone seed drafts.');
    } finally {
      setCreatingBulkDrafts(false);
    }
  };

  const handleBulkToggleZone = useCallback((zoneId: string) => {
    setBulkSelectedZoneIds((prev) => {
      if (prev.includes(zoneId)) {
        return prev.filter((existingId) => existingId !== zoneId);
      }
      return [...prev, zoneId];
    });
  }, []);

  const handleSelectMatchingZones = useCallback(() => {
    setBulkSelectedZoneIds((prev) => {
      const next = new Set(prev);
      bulkMatchingZones.forEach((zone) => next.add(zone.id));
      return Array.from(next);
    });
  }, [bulkMatchingZones]);

  const handleClearBulkSelection = useCallback(() => {
    setBulkSelectedZoneIds([]);
  }, []);

  const handleApprove = async (job: ZoneSeedJob) => {
    if (approvingId) return;
    setApprovingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post(`/sonar/admin/zone-seed-jobs/${job.id}/approve`);
      setSuccess('Draft approved and applying.');
      await fetchJobs(jobFilterZoneId || undefined, jobFilterStatuses);
    } catch (err) {
      console.error('Failed to approve draft', err);
      setError('Failed to approve draft.');
    } finally {
      setApprovingId(null);
    }
  };

  const handleDelete = async (job: ZoneSeedJob) => {
    if (deletingId) return;
    const confirmed = window.confirm(
      `Delete draft job ${job.id.slice(0, 8)}? This cannot be undone.`
    );
    if (!confirmed) return;
    setDeletingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.delete(`/sonar/admin/zone-seed-jobs/${job.id}`);
      setJobs((prev) => prev.filter((existing) => existing.id !== job.id));
      setSuccess('Draft job deleted.');
    } catch (err) {
      console.error('Failed to delete draft job', err);
      setError('Failed to delete draft job.');
    } finally {
      setDeletingId(null);
    }
  };

  const handleRetry = async (job: ZoneSeedJob) => {
    if (retryingId) return;
    setRetryingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post(`/sonar/admin/zone-seed-jobs/${job.id}/retry`);
      setSuccess('Draft retry queued.');
      await fetchJobs(jobFilterZoneId || undefined, jobFilterStatuses);
    } catch (err) {
      console.error('Failed to retry draft job', err);
      setError('Failed to retry draft job.');
    } finally {
      setRetryingId(null);
    }
  };

  const addRequiredTag = (value: string) => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (requiredPlaceTags.includes(trimmed)) return;
    setRequiredPlaceTags((prev) => [...prev, trimmed]);
  };

  const removeRequiredTag = (value: string) => {
    setRequiredPlaceTags((prev) => prev.filter((tag) => tag !== value));
  };

  const addShopkeeperTag = (value: string) => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (shopkeeperItemTags.includes(trimmed)) return;
    setShopkeeperItemTags((prev) => [...prev, trimmed]);
  };

  const removeShopkeeperTag = (value: string) => {
    setShopkeeperItemTags((prev) => prev.filter((tag) => tag !== value));
  };

  const filteredTagSuggestions = useMemo(() => {
    const query = requiredTagQuery.trim().toLowerCase();
    const available = knownPlaceTags.filter(
      (tag) => !requiredPlaceTags.includes(tag)
    );
    if (!query) return available;
    return available.filter((tag) => tag.includes(query));
  }, [knownPlaceTags, requiredPlaceTags, requiredTagQuery]);

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Zone Seeding</h1>
          <p className="text-sm text-gray-500">
            Create fantasy zone drafts with POIs, standalone challenges,
            scalable encounters, and treasure.
          </p>
        </div>
        <button
          className="px-4 py-2 rounded bg-gray-800 text-white hover:bg-gray-700"
          onClick={() =>
            fetchJobs(jobFilterZoneId || undefined, jobFilterStatuses)
          }
          disabled={loadingJobs}
        >
          {loadingJobs ? 'Refreshing...' : 'Refresh drafts'}
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-emerald-700">
          {success}
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            Draft settings
          </h2>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Zone
          </label>
          <div className="relative">
            <input
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
              value={draftZoneQuery}
              onChange={(e) => {
                const value = e.target.value;
                setDraftZoneQuery(value);
                setShowDraftZoneSuggestions(true);
                if (value.trim() === '') {
                  setDraftZoneId('');
                }
              }}
              onFocus={() => setShowDraftZoneSuggestions(true)}
              onBlur={() => {
                setTimeout(() => setShowDraftZoneSuggestions(false), 120);
              }}
              placeholder="Type to filter zones..."
            />
            {showDraftZoneSuggestions && draftZoneSuggestions.length > 0 && (
              <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                {draftZoneSuggestions.map((zone) => (
                  <button
                    type="button"
                    key={zone.id}
                    onClick={() => {
                      setDraftZoneId(zone.id);
                      setDraftZoneQuery(zone.name);
                      setShowDraftZoneSuggestions(false);
                    }}
                    className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                  >
                    {zone.name}
                  </button>
                ))}
              </div>
            )}
          </div>
          {selectedZone && (
            <p className="mt-2 text-xs text-gray-500">
              Selected: {selectedZone.name}
            </p>
          )}
          <div className="mt-4">
            <label className="block text-xs font-medium text-gray-500 mb-2">
              Seeding mode
            </label>
            <div className="inline-flex rounded border border-gray-300 bg-gray-50 p-1">
              <button
                type="button"
                className={`rounded px-3 py-1.5 text-sm font-medium ${
                  seedMode === 'manual'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
                onClick={() => setSeedMode('manual')}
              >
                Manual
              </button>
              <button
                type="button"
                className={`rounded px-3 py-1.5 text-sm font-medium ${
                  seedMode === 'auto'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
                onClick={() => setSeedMode('auto')}
              >
                Auto
              </button>
            </div>
            <p className="mt-2 text-xs text-gray-500">
              Auto mode uses the selected zone boundary area to recommend
              counts. Bulk queueing stays manual-only.
            </p>
          </div>
          {seedMode === 'auto' && (
            <div
              className={`mt-4 rounded border px-3 py-3 text-sm ${
                autoRecommendation
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-900'
                  : 'border-red-200 bg-red-50 text-red-700'
              }`}
            >
              {autoRecommendation ? (
                <>
                  <div className="font-medium">
                    Area-based recommendation active
                  </div>
                  <div className="mt-1 text-xs text-emerald-800">
                    Boundary area:{' '}
                    {formatSquareFeet(autoAreaSquareFeet ?? undefined)} (
                    {formatAcres(autoAreaAcres ?? undefined)}).
                  </div>
                  {autoFinalPreviewCounts && (
                    <div className="mt-1 text-xs text-emerald-800">
                      Current queued counts:{' '}
                      {formatZoneSeedCounts(autoFinalPreviewCounts)}.
                    </div>
                  )}
                  <div className="mt-1 text-xs text-emerald-800">
                    Final auto counts can still be adjusted at queue time if
                    place availability is lower than the recommendation.
                  </div>
                </>
              ) : (
                <>
                  <div className="font-medium">
                    Auto mode is unavailable for this zone
                  </div>
                  <div className="mt-1 text-xs text-red-700">
                    This zone needs valid boundary geometry before auto
                    inference can be queued.
                  </div>
                </>
              )}
            </div>
          )}
          <div className="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
            {seedCountFields.map((field) => {
              const inputValue =
                seedMode === 'auto'
                  ? autoDisplayedCountInputs[field.key]
                  : manualCountInputs[field.key];
              const recommendedValue = autoRecommendation?.[field.key];
              const hasOverride =
                seedMode === 'auto' &&
                autoOverrideInputs[field.key] != null &&
                autoOverrideInputs[field.key] !==
                  String(recommendedValue ?? '');

              return (
                <div key={field.key}>
                  <label className="block text-xs font-medium text-gray-500 mb-1">
                    {field.label}
                  </label>
                  <input
                    className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                    value={inputValue}
                    onChange={(e) =>
                      seedMode === 'auto'
                        ? updateAutoOverrideInput(field.key, e.target.value)
                        : updateManualCountInput(field.key, e.target.value)
                    }
                  />
                  <p className="mt-1 text-[11px] text-gray-500">
                    {field.description}
                    {seedMode === 'auto' && typeof recommendedValue === 'number'
                      ? hasOverride
                        ? ` Recommended: ${recommendedValue}.`
                        : ` Recommended: ${recommendedValue}.`
                      : ''}
                  </p>
                </div>
              );
            })}
          </div>
          <div className="mt-4">
            <label className="block text-xs font-medium text-gray-500 mb-1">
              Required POI tags
            </label>
            <div className="rounded border border-gray-300 px-2 py-2 text-sm">
              <div className="flex flex-wrap gap-2">
                {requiredPlaceTags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-indigo-50 px-2 py-1 text-xs text-indigo-700"
                  >
                    {tag}
                    <button
                      type="button"
                      className="ml-2 text-indigo-500 hover:text-indigo-700"
                      onClick={() => removeRequiredTag(tag)}
                    >
                      x
                    </button>
                  </span>
                ))}
                <div className="relative flex-1 min-w-[140px]">
                  <input
                    className="w-full border-0 px-2 py-1 text-sm focus:outline-none"
                    placeholder="Add tag..."
                    value={requiredTagQuery}
                    onChange={(e) => {
                      setRequiredTagQuery(e.target.value);
                      setShowRequiredTagSuggestions(true);
                    }}
                    onFocus={() => setShowRequiredTagSuggestions(true)}
                    onBlur={() =>
                      setTimeout(
                        () => setShowRequiredTagSuggestions(false),
                        120
                      )
                    }
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ',') {
                        e.preventDefault();
                        addRequiredTag(requiredTagQuery);
                        setRequiredTagQuery('');
                      }
                      if (
                        e.key === 'Backspace' &&
                        requiredTagQuery === '' &&
                        requiredPlaceTags.length > 0
                      ) {
                        removeRequiredTag(
                          requiredPlaceTags[requiredPlaceTags.length - 1]
                        );
                      }
                    }}
                  />
                  {showRequiredTagSuggestions &&
                    (filteredTagSuggestions.length > 0 ||
                      requiredTagQuery.trim()) && (
                      <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                        {filteredTagSuggestions.map((tag) => (
                          <button
                            key={tag}
                            type="button"
                            className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                            onClick={() => {
                              addRequiredTag(tag);
                              setRequiredTagQuery('');
                              setShowRequiredTagSuggestions(false);
                            }}
                          >
                            {tag}
                          </button>
                        ))}
                        {requiredTagQuery.trim() &&
                          !requiredPlaceTags.includes(
                            requiredTagQuery.trim().toLowerCase()
                          ) && (
                            <button
                              type="button"
                              className="block w-full px-3 py-2 text-left text-sm text-indigo-600 hover:bg-indigo-50"
                              onClick={() => {
                                addRequiredTag(requiredTagQuery);
                                setRequiredTagQuery('');
                                setShowRequiredTagSuggestions(false);
                              }}
                            >
                              Add &quot;{requiredTagQuery.trim()}&quot;
                            </button>
                          )}
                      </div>
                    )}
                </div>
              </div>
            </div>
            <p className="mt-1 text-xs text-gray-400">
              We will ensure at least one POI matches each tag.
            </p>
          </div>
          <div className="mt-4">
            <label className="block text-xs font-medium text-gray-500 mb-1">
              Shopkeeper item tags
            </label>
            <div className="rounded border border-gray-300 px-2 py-2 text-sm">
              <div className="flex flex-wrap gap-2">
                {shopkeeperItemTags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-emerald-50 px-2 py-1 text-xs text-emerald-700"
                  >
                    {tag}
                    <button
                      type="button"
                      className="ml-2 text-emerald-600 hover:text-emerald-800"
                      onClick={() => removeShopkeeperTag(tag)}
                    >
                      x
                    </button>
                  </span>
                ))}
                <input
                  className="flex-1 min-w-[140px] border-0 px-2 py-1 text-sm focus:outline-none"
                  placeholder="Add item tag..."
                  value={shopkeeperTagQuery}
                  onChange={(e) => setShopkeeperTagQuery(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ',') {
                      e.preventDefault();
                      addShopkeeperTag(shopkeeperTagQuery);
                      setShopkeeperTagQuery('');
                    }
                    if (
                      e.key === 'Backspace' &&
                      shopkeeperTagQuery === '' &&
                      shopkeeperItemTags.length > 0
                    ) {
                      removeShopkeeperTag(
                        shopkeeperItemTags[shopkeeperItemTags.length - 1]
                      );
                    }
                  }}
                />
              </div>
            </div>
            <p className="mt-1 text-xs text-gray-400">
              One shopkeeper will be generated per tag at a random location in
              the zone. Auto mode does not infer these tags for you.
            </p>
          </div>
          <button
            className="mt-5 w-full rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-500 disabled:opacity-60"
            onClick={handleCreateDraft}
            disabled={
              creatingDraft ||
              !draftZoneId ||
              (seedMode === 'auto' && !autoRecommendation)
            }
          >
            {creatingDraft ? 'Queuing...' : 'Create draft'}
          </button>
          <div className="mt-6 rounded-lg border border-dashed border-gray-300 bg-gray-50 p-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-gray-900">
                  Bulk queue
                </h3>
                <p className="text-xs text-gray-500">
                  Queue this same seed configuration across many zones at once.
                </p>
                {seedMode === 'auto' && (
                  <p className="mt-1 text-xs text-amber-700">
                    Switch back to manual mode to bulk queue drafts.
                  </p>
                )}
              </div>
            </div>
            <div className="mt-3">
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Zone filter
              </label>
              <input
                className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                value={bulkZoneQuery}
                onChange={(e) => setBulkZoneQuery(e.target.value)}
                placeholder="Optional name filter..."
              />
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <button
                type="button"
                className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                onClick={handleSelectMatchingZones}
                disabled={bulkMatchingZones.length === 0}
              >
                Select matching
              </button>
              <button
                type="button"
                className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                onClick={handleClearBulkSelection}
                disabled={bulkTargetZones.length === 0}
              >
                Clear selection
              </button>
            </div>
            <BulkZoneSelectionMap
              zones={sortedZones}
              matchingZoneIds={bulkMatchingZoneIds}
              selectedZoneIds={bulkSelectedZoneIdSet}
              onToggleZone={handleBulkToggleZone}
            />
            <p className="mt-2 text-xs text-gray-500">
              Matching zones: {bulkMatchingZones.length}. Selected zones:{' '}
              {bulkTargetZones.length}.
            </p>
            {bulkTargetZones.length > 0 && (
              <p className="mt-2 text-xs text-gray-500">
                Selected:{' '}
                {bulkTargetZones
                  .slice(0, 5)
                  .map((zone) => zone.name)
                  .join(', ')}
                {bulkTargetZones.length > 5
                  ? ` +${bulkTargetZones.length - 5} more`
                  : ''}
              </p>
            )}
            <button
              className="mt-4 w-full rounded bg-slate-800 px-4 py-2 text-white hover:bg-slate-700 disabled:opacity-60"
              onClick={handleBulkCreateDrafts}
              disabled={
                creatingBulkDrafts ||
                bulkTargetZones.length === 0 ||
                seedMode !== 'manual'
              }
            >
              {creatingBulkDrafts
                ? 'Queuing bulk drafts...'
                : `Queue for ${bulkTargetZones.length} zones`}
            </button>
          </div>
        </div>

        <div className="lg:col-span-2 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <div className="mb-4 space-y-3">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <h2 className="text-lg font-semibold text-gray-900">
                Draft jobs
              </h2>
              <div className="relative w-full md:w-72">
                <input
                  className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  value={filterZoneQuery}
                  onChange={(e) => {
                    const value = e.target.value;
                    setFilterZoneQuery(value);
                    setShowFilterZoneSuggestions(true);
                    if (value.trim() === '') {
                      setJobFilterZoneId('');
                    }
                  }}
                  onFocus={() => setShowFilterZoneSuggestions(true)}
                  onBlur={() =>
                    setTimeout(() => setShowFilterZoneSuggestions(false), 120)
                  }
                  placeholder="Filter by zone (optional)..."
                />
                {showFilterZoneSuggestions &&
                  filterZoneSuggestions.length > 0 && (
                    <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                      <button
                        type="button"
                        onClick={() => {
                          setJobFilterZoneId('');
                          setFilterZoneQuery('');
                          setShowFilterZoneSuggestions(false);
                        }}
                        className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                      >
                        All zones
                      </button>
                      {filterZoneSuggestions.map((zone) => (
                        <button
                          type="button"
                          key={zone.id}
                          onClick={() => {
                            setJobFilterZoneId(zone.id);
                            setFilterZoneQuery(zone.name);
                            setShowFilterZoneSuggestions(false);
                          }}
                          className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                        >
                          {zone.name}
                        </button>
                      ))}
                    </div>
                  )}
              </div>
            </div>

            <div>
              <div className="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500">
                State filters
              </div>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => setJobFilterStatuses([])}
                  className={`rounded-full px-3 py-1 text-xs font-medium ${
                    jobFilterStatuses.length === 0
                      ? 'bg-gray-900 text-white'
                      : 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  All states
                </button>
                {zoneSeedJobStateOptions.map((status) => {
                  const active = jobFilterStatuses.includes(status);
                  return (
                    <button
                      key={status}
                      type="button"
                      onClick={() => toggleJobFilterStatus(status)}
                      className={`rounded-full px-3 py-1 text-xs font-medium ${
                        active
                          ? 'bg-indigo-600 text-white'
                          : 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
                      }`}
                    >
                      {status.replace(/_/g, ' ')}
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 rounded border border-gray-200 bg-gray-50 px-3 py-2">
              <div className="text-sm text-gray-600">
                {selectedJobIds.size > 0
                  ? `${selectedJobIds.size} draft job${selectedJobIds.size === 1 ? '' : 's'} selected`
                  : `${visibleDeletableJobs.length} visible deletable draft job${visibleDeletableJobs.length === 1 ? '' : 's'}`}
              </div>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  onClick={handleSelectVisibleJobs}
                  disabled={
                    visibleDeletableJobs.length === 0 ||
                    allVisibleDeletableJobsSelected
                  }
                >
                  {allVisibleDeletableJobsSelected
                    ? 'All visible selected'
                    : 'Select visible'}
                </button>
                <button
                  type="button"
                  className="rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  onClick={clearJobSelection}
                  disabled={selectedJobIds.size === 0}
                >
                  Clear selection
                </button>
                <button
                  type="button"
                  className="rounded bg-red-600 px-3 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
                  onClick={handleBulkDeleteJobs}
                  disabled={
                    selectedJobIds.size === 0 ||
                    bulkDeletingJobs ||
                    deletingId !== null
                  }
                >
                  {bulkDeletingJobs
                    ? 'Deleting selected...'
                    : 'Delete selected'}
                </button>
              </div>
            </div>
          </div>
          {loadingJobs ? (
            <p className="text-sm text-gray-500">Loading drafts...</p>
          ) : jobs.length === 0 ? (
            <p className="text-sm text-gray-500">
              {jobFilterZoneId || jobFilterStatuses.length > 0
                ? 'No draft jobs match these filters.'
                : 'No draft jobs for this zone yet.'}
            </p>
          ) : (
            <div className="space-y-4">
              {jobs.map((job) => {
                const zoneName = zoneNameById.get(job.zoneId) || job.zoneId;
                const canDelete = canDeleteZoneSeedJob(job);
                const finalCounts = getJobFinalCounts(job);
                const recommendedCounts = job.autoSeedAudit?.recommendedCounts;
                return (
                  <div
                    key={job.id}
                    className="rounded-lg border border-gray-200 p-4"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-start gap-3">
                        <input
                          type="checkbox"
                          checked={selectedJobIds.has(job.id)}
                          disabled={!canDelete || bulkDeletingJobs}
                          onChange={() => handleToggleJobSelection(job.id)}
                          className="mt-1 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 disabled:cursor-not-allowed disabled:opacity-40"
                          title={
                            canDelete
                              ? 'Select draft job'
                              : 'Cannot select a running job'
                          }
                        />
                        <div>
                          <h3 className="text-sm font-semibold text-gray-900">
                            Job {job.id.slice(0, 8)}
                          </h3>
                          <p className="text-xs font-medium text-indigo-700">
                            Zone: {zoneName}
                          </p>
                          <p className="text-xs text-gray-500">
                            Created: {formatDate(job.createdAt)} | Updated:{' '}
                            {formatDate(job.updatedAt)}
                          </p>
                          <p className="text-xs text-gray-500">
                            Mode: {job.seedMode || 'manual'} | Counts:{' '}
                            {formatZoneSeedCounts(finalCounts)},{' '}
                            {job.shopkeeperItemTags?.length ?? 0} shopkeepers
                          </p>
                          {job.seedMode === 'auto' && job.autoSeedAudit && (
                            <p className="text-xs text-gray-500">
                              Auto audit:{' '}
                              {formatSquareFeet(
                                job.autoSeedAudit.zoneAreaSquareFeet
                              )}{' '}
                              ({formatAcres(job.autoSeedAudit.zoneAreaAcres)})
                              {recommendedCounts
                                ? ` | Recommended: ${formatZoneSeedCounts(recommendedCounts)}`
                                : ''}
                            </p>
                          )}
                          {job.requiredPlaceTags &&
                            job.requiredPlaceTags.length > 0 && (
                              <p className="text-xs text-gray-500">
                                Required tags:{' '}
                                {job.requiredPlaceTags.join(', ')}
                              </p>
                            )}
                          {job.shopkeeperItemTags &&
                            job.shopkeeperItemTags.length > 0 && (
                              <p className="text-xs text-gray-500">
                                Shopkeeper tags:{' '}
                                {job.shopkeeperItemTags.join(', ')}
                              </p>
                            )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <span
                          className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold text-white ${statusBadgeClass(
                            job.status
                          )}`}
                        >
                          {job.status.replace(/_/g, ' ')}
                        </span>
                        {job.status === 'failed' && (
                          <button
                            className="rounded border border-gray-200 px-2 py-1 text-xs text-indigo-700 hover:bg-indigo-50 disabled:opacity-50"
                            onClick={() => handleRetry(job)}
                            disabled={retryingId === job.id}
                            title="Retry draft job"
                          >
                            {retryingId === job.id ? 'Retrying...' : 'Retry'}
                          </button>
                        )}
                        <button
                          className="rounded border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-gray-50 disabled:opacity-50"
                          onClick={() => handleDelete(job)}
                          disabled={
                            deletingId === job.id ||
                            bulkDeletingJobs ||
                            !canDelete
                          }
                          title={
                            !canDelete
                              ? 'Cannot delete while running'
                              : 'Delete draft job'
                          }
                        >
                          {deletingId === job.id ? 'Deleting...' : 'Delete'}
                        </button>
                      </div>
                    </div>

                    {job.errorMessage && (
                      <div className="mt-3 rounded border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-700">
                        {job.errorMessage}
                      </div>
                    )}

                    {job.seedMode === 'auto' &&
                      job.autoSeedAudit?.warnings &&
                      job.autoSeedAudit.warnings.length > 0 && (
                        <div className="mt-3 rounded border border-amber-100 bg-amber-50 px-3 py-2 text-xs text-amber-800">
                          {job.autoSeedAudit.warnings.join(' ')}
                        </div>
                      )}

                    {job.draft && (
                      <details className="mt-3">
                        <summary className="cursor-pointer text-sm font-medium text-gray-700">
                          Draft details
                        </summary>
                        <div className="mt-3 space-y-6 text-sm text-gray-700">
                          <div>
                            <div className="font-semibold">
                              Fantasy branding
                            </div>
                            <div className="text-sm text-gray-600">
                              {job.draft.fantasyName || 'Untitled district'}
                            </div>
                            {job.draft.zoneDescription && (
                              <p className="mt-2 text-sm text-gray-600 whitespace-pre-wrap">
                                {job.draft.zoneDescription}
                              </p>
                            )}
                          </div>
                          <div>
                            <div className="font-semibold">
                              Points of interest
                            </div>
                            <div className="mt-2 space-y-3 text-xs text-gray-600">
                              {(job.draft.pointsOfInterest || []).map((poi) => (
                                <div
                                  key={poi.draftId}
                                  className="rounded border border-gray-100 bg-gray-50 p-3"
                                >
                                  <div className="text-sm font-semibold text-gray-800">
                                    {poi.name || 'Unnamed place'}
                                  </div>
                                  <div>Place ID: {poi.placeId || 'n/a'}</div>
                                  {poi.address && (
                                    <div>Address: {poi.address}</div>
                                  )}
                                  {typeof poi.latitude === 'number' &&
                                    typeof poi.longitude === 'number' && (
                                      <div>
                                        Coordinates: {poi.latitude},{' '}
                                        {poi.longitude}
                                      </div>
                                    )}
                                  {typeof poi.rating === 'number' && (
                                    <div>
                                      Rating: {poi.rating}
                                      {typeof poi.userRatingCount === 'number'
                                        ? ` (${poi.userRatingCount} reviews)`
                                        : ''}
                                    </div>
                                  )}
                                  {poi.types && poi.types.length > 0 && (
                                    <div>Types: {poi.types.join(', ')}</div>
                                  )}
                                  {poi.editorialSummary && (
                                    <div className="mt-1 text-gray-500">
                                      Summary: {poi.editorialSummary}
                                    </div>
                                  )}
                                </div>
                              ))}
                            </div>
                          </div>
                          <div>
                            <div className="font-semibold">Characters</div>
                            <div className="mt-2 space-y-3 text-xs text-gray-600">
                              {(job.draft.characters || []).map((character) => (
                                <div
                                  key={character.draftId}
                                  className="rounded border border-gray-100 bg-gray-50 p-3"
                                >
                                  <div className="text-sm font-semibold text-gray-800">
                                    {character.name || 'Unnamed character'}
                                  </div>
                                  <div>
                                    Place ID: {character.placeId || 'n/a'}
                                  </div>
                                  {typeof character.latitude === 'number' &&
                                    typeof character.longitude === 'number' && (
                                      <div>
                                        Coordinates: {character.latitude},{' '}
                                        {character.longitude}
                                      </div>
                                    )}
                                  {character.shopItemTags &&
                                    character.shopItemTags.length > 0 && (
                                      <div>
                                        Shopkeeper tags:{' '}
                                        {character.shopItemTags.join(', ')}
                                      </div>
                                    )}
                                  {character.description && (
                                    <div className="mt-1 text-gray-500 whitespace-pre-wrap">
                                      {character.description}
                                    </div>
                                  )}
                                </div>
                              ))}
                            </div>
                          </div>
                          <div>
                            {job.seedMode === 'auto' && job.autoSeedAudit && (
                              <div className="mb-6">
                                <div className="font-semibold">
                                  Auto seed audit
                                </div>
                                <div className="mt-2 rounded border border-emerald-100 bg-emerald-50 p-3 text-xs text-emerald-900">
                                  <div>
                                    Boundary area:{' '}
                                    {formatSquareFeet(
                                      job.autoSeedAudit.zoneAreaSquareFeet
                                    )}{' '}
                                    (
                                    {formatAcres(
                                      job.autoSeedAudit.zoneAreaAcres
                                    )}
                                    )
                                  </div>
                                  {recommendedCounts && (
                                    <div className="mt-1">
                                      Recommended counts:{' '}
                                      {formatZoneSeedCounts(recommendedCounts)}
                                    </div>
                                  )}
                                  {job.autoSeedAudit.finalCounts && (
                                    <div className="mt-1">
                                      Final queued counts:{' '}
                                      {formatZoneSeedCounts(
                                        job.autoSeedAudit.finalCounts
                                      )}
                                    </div>
                                  )}
                                  {job.autoSeedAudit.warnings &&
                                    job.autoSeedAudit.warnings.length > 0 && (
                                      <div className="mt-2 space-y-1 text-amber-900">
                                        {job.autoSeedAudit.warnings.map(
                                          (warning, index) => (
                                            <div
                                              key={`${job.id}-warning-${index}`}
                                            >
                                              {warning}
                                            </div>
                                          )
                                        )}
                                      </div>
                                    )}
                                </div>
                              </div>
                            )}
                          </div>
                          <div>
                            <div className="font-semibold">
                              Seeding plan preview
                            </div>
                            <div className="mt-2 rounded border border-gray-100 bg-gray-50 p-3 text-xs text-gray-600">
                              <div>
                                {job.placeCount} POIs selected for challenge
                                placement
                              </div>
                              <div>
                                {job.placeCount} standalone challenges at those
                                POIs
                              </div>
                              <div>
                                {job.monsterCount ?? 0} random monster
                                encounters (scalable)
                              </div>
                              <div>
                                {job.bossEncounterCount ?? 0} random boss
                                encounters (scalable +5 levels)
                              </div>
                              <div>
                                {job.raidEncounterCount ?? 0} random raid
                                encounters (scaled for 5-player parties)
                              </div>
                              <div>
                                {job.inputEncounterCount ?? 0} random input
                                scenarios (scalable)
                              </div>
                              <div>
                                {job.optionEncounterCount ?? 0} random option
                                scenarios (scalable)
                              </div>
                              <div>
                                {job.treasureChestCount ?? 0} random treasure
                                chests (scalable rewards)
                              </div>
                              <div>
                                {job.healingFountainCount ?? 0} random healing
                                fountains
                              </div>
                              <div>
                                {job.shopkeeperItemTags?.length ?? 0}{' '}
                                shopkeepers generated at random zone locations
                              </div>
                            </div>
                          </div>
                        </div>
                      </details>
                    )}

                    {job.status === 'awaiting_approval' && (
                      <div className="mt-4">
                        <button
                          className="rounded bg-emerald-600 px-4 py-2 text-sm text-white hover:bg-emerald-500 disabled:opacity-60"
                          onClick={() => handleApprove(job)}
                          disabled={approvingId === job.id}
                        >
                          {approvingId === job.id
                            ? 'Approving...'
                            : 'Approve and apply'}
                        </button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ZoneSeedJobs;
