import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone, ZoneAdminSummary, ZoneImport } from '@poltergeist/types';
import { useNavigate } from 'react-router-dom';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import * as wellknown from 'wellknown';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

const parseInternalTagsInput = (value: string): string[] =>
  Array.from(
    new Set(
      value
        .split(',')
        .map((tag) => tag.trim().toLowerCase())
        .filter((tag) => tag !== '')
    )
  );

const parseBoundaryString = (boundary: string): [number, number][] => {
  const trimmed = boundary.trim();
  if (!trimmed) return [];

  if (
    trimmed.toUpperCase().startsWith('POLYGON') ||
    trimmed.startsWith('SRID=')
  ) {
    const wkt = trimmed.includes(';')
      ? trimmed.split(';').slice(1).join(';').trim()
      : trimmed;
    const parsed = wellknown.parse(wkt) as
      | GeoJSON.Polygon
      | GeoJSON.MultiPolygon
      | null;
    if (!parsed) return [];
    if (parsed.type === 'Polygon') {
      return parsed.coordinates?.[0] as [number, number][];
    }
    if (parsed.type === 'MultiPolygon') {
      return parsed.coordinates?.[0]?.[0] as [number, number][];
    }
    return [];
  }

  if (trimmed.startsWith('[')) {
    try {
      const parsed = JSON.parse(trimmed);
      if (Array.isArray(parsed)) {
        return parsed
          .filter((pair) => Array.isArray(pair) && pair.length >= 2)
          .map((pair) => [Number(pair[0]), Number(pair[1])] as [number, number])
          .filter((pair) => !Number.isNaN(pair[0]) && !Number.isNaN(pair[1]));
      }
    } catch (error) {
      console.warn('Failed to parse zone boundary JSON', error);
    }
  }

  return [];
};

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

  const boundaryUnknown = zone.boundary as unknown;
  if (typeof boundaryUnknown === 'string') {
    return sortBoundaryPoints(parseBoundaryString(boundaryUnknown));
  }
  if (Array.isArray(boundaryUnknown)) {
    const coords = boundaryUnknown
      .filter((pair) => Array.isArray(pair) && pair.length >= 2)
      .map((pair) => [Number(pair[0]), Number(pair[1])] as [number, number])
      .filter((pair) => !Number.isNaN(pair[0]) && !Number.isNaN(pair[1]));
    return sortBoundaryPoints(coords);
  }

  return [];
};

const calculateBoundaryCenter = (points: [number, number][]) => {
  const openPoints =
    points.length > 1 &&
    points[0][0] === points[points.length - 1][0] &&
    points[0][1] === points[points.length - 1][1]
      ? points.slice(0, -1)
      : points;

  if (openPoints.length === 0) {
    return null;
  }

  const totals = openPoints.reduce(
    (acc, point) => [acc[0] + point[0], acc[1] + point[1]],
    [0, 0]
  );

  return {
    longitude: totals[0] / openPoints.length,
    latitude: totals[1] / openPoints.length,
  };
};

const formatCoordinate = (value: number) => value.toFixed(4);

const formatZoneDate = (value: string) =>
  new Date(value).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

const buildZoneSearchText = (zone: ZoneAdminSummary) =>
  [
    zone.name,
    zone.description,
    zone.importMetroName || '',
    ...(zone.internalTags || []),
  ]
    .join(' ')
    .toLowerCase();

const getZoneActivityLabel = (zone: ZoneAdminSummary) => {
  const totalContent =
    zone.questCount +
    zone.zoneQuestArchetypeCount +
    zone.challengeCount +
    zone.scenarioCount +
    zone.monsterEncounterCount +
    zone.pointOfInterestCount +
    zone.treasureChestCount +
    zone.healingFountainCount;

  if (totalContent >= 30) {
    return 'Dense';
  }
  if (totalContent >= 15) {
    return 'Active';
  }
  if (totalContent >= 6) {
    return 'Growing';
  }
  if (totalContent >= 1) {
    return 'Light';
  }
  return 'Empty';
};

const getZoneReadinessSummary = (zone: ZoneAdminSummary) => {
  const objectiveCount =
    zone.challengeCount + zone.scenarioCount + zone.monsterEncounterCount;

  if (zone.questCount > 0 && objectiveCount > 0) {
    return `${zone.questCount} live quest${zone.questCount === 1 ? '' : 's'} supported by ${objectiveCount} linked objective${objectiveCount === 1 ? '' : 's'}.`;
  }

  if (zone.zoneQuestArchetypeCount > 0) {
    return `${zone.zoneQuestArchetypeCount} assigned archetype${zone.zoneQuestArchetypeCount === 1 ? '' : 's'} ready for generation, with ${zone.pointOfInterestCount} mapped point${zone.pointOfInterestCount === 1 ? '' : 's'} of interest.`;
  }

  if (zone.pointOfInterestCount > 0) {
    return `${zone.pointOfInterestCount} mapped point${zone.pointOfInterestCount === 1 ? '' : 's'} of interest and ${zone.boundaryPointCount} boundary point${zone.boundaryPointCount === 1 ? '' : 's'} are in place, but quest content is still thin.`;
  }

  return 'This zone has geometry, but it still needs tags, POIs, and quest content to feel fully authored.';
};

type BoundaryEditorMapProps = {
  center: [number, number];
  boundaryPoints: [number, number][];
  allZoneBoundaries: [number, number][][];
  onMapClick: (lngLat: mapboxgl.LngLat) => void;
  onClearBoundary: () => void;
};

const ZoneBoundaryEditorMap: React.FC<BoundaryEditorMapProps> = ({
  center,
  boundaryPoints,
  allZoneBoundaries,
  onMapClick,
  onClearBoundary,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const markers = useRef<mapboxgl.Marker[]>([]);
  const searchMarker = useRef<mapboxgl.Marker | null>(null);
  const [isLocating, setIsLocating] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [showSearchSuggestions, setShowSearchSuggestions] = useState(false);
  const [searchCandidates, setSearchCandidates] = useState<
    Array<{
      id: string;
      center: [number, number];
      placeName: string;
    }>
  >([]);
  const [locationError, setLocationError] = useState<string | null>(null);
  const [searchStatus, setSearchStatus] = useState<string | null>(null);

  const focusMapOnLocation = (
    lngLat: [number, number],
    zoom: number,
    markerColor = '#2563EB'
  ) => {
    if (!map.current) {
      return;
    }

    searchMarker.current?.remove();
    searchMarker.current = new mapboxgl.Marker({ color: markerColor })
      .setLngLat(lngLat)
      .addTo(map.current);

    map.current.flyTo({
      center: lngLat,
      zoom: Math.max(map.current.getZoom(), zoom),
      essential: true,
    });
  };

  const handleSearchLocation = async () => {
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
    setLocationError(null);
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
      setSearchStatus('Select a location below to move the map.');
    } catch (error) {
      console.error('Error searching for location:', error);
      setSearchCandidates([]);
      setSearchStatus('Unable to search for that location right now.');
    } finally {
      setIsSearching(false);
    }
  };

  const cleanupBoundary = () => {
    if (!map.current) {
      return;
    }
    if (map.current.getLayer('zone-create-boundary-outline')) {
      map.current.removeLayer('zone-create-boundary-outline');
    }
    if (map.current.getLayer('zone-create-boundary')) {
      map.current.removeLayer('zone-create-boundary');
    }
    if (map.current.getSource('zone-create-boundary')) {
      map.current.removeSource('zone-create-boundary');
    }
  };

  useEffect(() => {
    if (mapContainer.current && !map.current) {
      map.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center,
        zoom: 12,
        interactive: true,
      });

      map.current.on('load', () => {
        setMapLoaded(true);
      });
    }

    return () => {
      cleanupBoundary();
      searchMarker.current?.remove();
      searchMarker.current = null;
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    map.current.flyTo({
      center,
      essential: true,
    });
  }, [center, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    const canvasContainer = mapContainer.current?.querySelector(
      '.mapboxgl-canvas-container'
    );
    if (canvasContainer instanceof HTMLElement) {
      canvasContainer.style.cursor = 'grab';
    }

    const handleClick = (event: mapboxgl.MapMouseEvent) => {
      onMapClick(event.lngLat);
    };

    map.current.on('click', handleClick);
    return () => {
      map.current?.off('click', handleClick);
    };
  }, [mapLoaded, onMapClick]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    markers.current.forEach((marker) => marker.remove());
    markers.current = [];

    const sortedPoints = sortBoundaryPoints(boundaryPoints);
    sortedPoints.forEach((point) => {
      const marker = new mapboxgl.Marker({
        color: '#DC2626',
        draggable: false,
      })
        .setLngLat(point)
        .addTo(map.current!);
      markers.current.push(marker);
    });
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    cleanupBoundary();

    const sortedPoints = sortBoundaryPoints(boundaryPoints);
    if (sortedPoints.length < 4) {
      return;
    }

    map.current.addSource('zone-create-boundary', {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Polygon',
          coordinates: [sortedPoints],
        },
      },
    });

    map.current.addLayer({
      id: 'zone-create-boundary',
      type: 'fill',
      source: 'zone-create-boundary',
      paint: {
        'fill-color': '#DC2626',
        'fill-opacity': 0.3,
      },
    });

    map.current.addLayer({
      id: 'zone-create-boundary-outline',
      type: 'line',
      source: 'zone-create-boundary',
      paint: {
        'line-color': '#DC2626',
        'line-width': 2,
      },
    });
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    if (map.current.getLayer('zone-create-all-boundaries-outline')) {
      map.current.removeLayer('zone-create-all-boundaries-outline');
    }
    if (map.current.getLayer('zone-create-all-boundaries')) {
      map.current.removeLayer('zone-create-all-boundaries');
    }
    if (map.current.getSource('zone-create-all-boundaries')) {
      map.current.removeSource('zone-create-all-boundaries');
    }

    const polygonFeatures = allZoneBoundaries
      .map((points) => sortBoundaryPoints(points))
      .filter((points) => points.length >= 4)
      .map((points) => ({
        type: 'Feature' as const,
        properties: {},
        geometry: {
          type: 'Polygon' as const,
          coordinates: [points],
        },
      }));

    if (polygonFeatures.length === 0) {
      return;
    }

    map.current.addSource('zone-create-all-boundaries', {
      type: 'geojson',
      data: {
        type: 'FeatureCollection',
        features: polygonFeatures,
      },
    });

    map.current.addLayer(
      {
        id: 'zone-create-all-boundaries',
        type: 'fill',
        source: 'zone-create-all-boundaries',
        paint: {
          'fill-color': '#2563EB',
          'fill-opacity': 0.18,
        },
      },
      map.current.getLayer('zone-create-boundary')
        ? 'zone-create-boundary'
        : undefined
    );

    map.current.addLayer(
      {
        id: 'zone-create-all-boundaries-outline',
        type: 'line',
        source: 'zone-create-all-boundaries',
        paint: {
          'line-color': '#2563EB',
          'line-width': 2,
        },
      },
      map.current.getLayer('zone-create-boundary-outline')
        ? 'zone-create-boundary-outline'
        : undefined
    );

    return () => {
      if (!map.current) {
        return;
      }
      if (map.current.getLayer('zone-create-all-boundaries-outline')) {
        map.current.removeLayer('zone-create-all-boundaries-outline');
      }
      if (map.current.getLayer('zone-create-all-boundaries')) {
        map.current.removeLayer('zone-create-all-boundaries');
      }
      if (map.current.getSource('zone-create-all-boundaries')) {
        map.current.removeSource('zone-create-all-boundaries');
      }
    };
  }, [allZoneBoundaries, mapLoaded, boundaryPoints]);

  return (
    <div style={{ marginBottom: '15px' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '10px',
        }}
      >
        <div>
          <div style={{ fontWeight: 600, marginBottom: '4px' }}>Boundary</div>
          <div style={{ fontSize: '12px', color: '#6b7280' }}>
            Click on the map to add boundary points for the new zone.
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            type="button"
            onClick={onClearBoundary}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            Clear Boundary
          </button>
          <button
            type="button"
            onClick={() => map.current?.zoomIn()}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            +
          </button>
          <button
            type="button"
            onClick={() => map.current?.zoomOut()}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
          >
            −
          </button>
          <button
            type="button"
            onClick={() => {
              if (!navigator.geolocation) {
                setLocationError(
                  'Geolocation is not supported in this browser.'
                );
                return;
              }
              setIsLocating(true);
              setLocationError(null);
              navigator.geolocation.getCurrentPosition(
                (pos) => {
                  setIsLocating(false);
                  const { latitude, longitude } = pos.coords;
                  focusMapOnLocation([longitude, latitude], 16, '#16A34A');
                  setSearchCandidates([]);
                  setShowSearchSuggestions(false);
                  setSearchStatus('Moved map to your current location.');
                },
                (err) => {
                  setIsLocating(false);
                  setLocationError(err.message || 'Unable to fetch location.');
                },
                { enableHighAccuracy: true, timeout: 10000, maximumAge: 10000 }
              );
            }}
            style={{
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
              backgroundColor: '#fff',
            }}
            disabled={isLocating}
          >
            {isLocating ? '...' : 'My Location'}
          </button>
        </div>
      </div>
      <div
        style={{
          display: 'flex',
          gap: '8px',
          marginBottom: '10px',
          flexWrap: 'wrap',
        }}
      >
        <div
          style={{ position: 'relative', flex: '1 1 320px', minWidth: '240px' }}
        >
          <input
            type="text"
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
            style={{
              width: '100%',
              padding: '8px 12px',
              borderRadius: '4px',
              border: '1px solid #ccc',
            }}
          />
          {showSearchSuggestions && searchCandidates.length > 0 && (
            <div
              style={{
                position: 'absolute',
                top: '100%',
                left: 0,
                right: 0,
                marginTop: '4px',
                maxHeight: '220px',
                overflowY: 'auto',
                border: '1px solid #ccc',
                borderRadius: '4px',
                backgroundColor: '#fff',
                zIndex: 20,
                boxShadow: '0 6px 20px rgba(0, 0, 0, 0.08)',
              }}
            >
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
                  style={{
                    display: 'block',
                    width: '100%',
                    padding: '8px 10px',
                    textAlign: 'left',
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                  }}
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
          style={{
            padding: '8px 12px',
            borderRadius: '4px',
            border: '1px solid #ccc',
            backgroundColor: '#fff',
            fontWeight: 600,
          }}
          disabled={isSearching}
        >
          {isSearching ? 'Searching...' : 'Search'}
        </button>
      </div>
      <div
        ref={mapContainer}
        style={{
          width: '100%',
          height: '360px',
          borderRadius: '8px',
          overflow: 'hidden',
          border: '1px solid #d1d5db',
        }}
      />
      <div style={{ marginTop: '10px', fontSize: '12px', color: '#6b7280' }}>
        {boundaryPoints.length} points added
        {boundaryPoints.length < 3
          ? '. Add at least 3 points to create the zone.'
          : '.'}
      </div>
      {locationError && (
        <div style={{ marginTop: '8px', color: '#dc2626', fontSize: '12px' }}>
          {locationError}
        </div>
      )}
      {searchStatus && (
        <div style={{ marginTop: '8px', color: '#475569', fontSize: '12px' }}>
          {searchStatus}
        </div>
      )}
    </div>
  );
};

export const Zones = () => {
  const { zones, refreshZones } = useZoneContext();
  const { apiClient } = useAPI();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [internalTagsInput, setInternalTagsInput] = useState('');
  const [showCreateZone, setShowCreateZone] = useState(false);
  const [createBoundaryPoints, setCreateBoundaryPoints] = useState<
    [number, number][]
  >([]);
  const [createMapCenter, setCreateMapCenter] = useState<[number, number]>([
    -87.6298, 41.8781,
  ]);
  const [createZoneError, setCreateZoneError] = useState<string | null>(null);
  const [isCreatingZone, setIsCreatingZone] = useState(false);
  const [selectedMetro, setSelectedMetro] = useState('Chicago, Illinois');
  const [customMetro, setCustomMetro] = useState('');
  const [importJobs, setImportJobs] = useState<ZoneImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const [importing, setImporting] = useState(false);
  const [deletingImportId, setDeletingImportId] = useState<string | null>(null);
  const [deletingZoneId, setDeletingZoneId] = useState<string | null>(null);
  const [zoneSummaries, setZoneSummaries] = useState<ZoneAdminSummary[]>([]);
  const [zoneSummariesLoading, setZoneSummariesLoading] = useState(false);
  const [zoneSummariesError, setZoneSummariesError] = useState<string | null>(
    null
  );
  const [zoneSearchQuery, setZoneSearchQuery] = useState('');
  const [zoneSort, setZoneSort] = useState<
    'richest' | 'quests' | 'updated' | 'name'
  >('richest');
  const [, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const navigate = useNavigate();
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const popupRef = useRef<mapboxgl.Popup | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const handlersAttachedRef = useRef(false);
  const fitBoundsRef = useRef(false);

  const fetchZoneSummaries = useCallback(async () => {
    setZoneSummariesLoading(true);
    setZoneSummariesError(null);
    try {
      const response = await apiClient.get<ZoneAdminSummary[]>('/sonar/admin/zones');
      setZoneSummaries(response);
    } catch (error) {
      console.error('Failed to fetch zone summaries', error);
      setZoneSummariesError('Failed to load zone summaries.');
    } finally {
      setZoneSummariesLoading(false);
    }
  }, [apiClient]);

  const mostRecentlyCreatedZone = zones.reduce<Zone | null>((latest, zone) => {
    if (!latest) {
      return zone;
    }

    return new Date(zone.createdAt).getTime() >
      new Date(latest.createdAt).getTime()
      ? zone
      : latest;
  }, null);

  const zoneSummaryByID = useMemo(
    () => new Map(zoneSummaries.map((zone) => [zone.id, zone])),
    [zoneSummaries]
  );

  const zoneOverview = useMemo(
    () =>
      zoneSummaries.reduce(
        (totals, zone) => {
          totals.pointOfInterestCount += zone.pointOfInterestCount;
          totals.questCount += zone.questCount;
          totals.zoneQuestArchetypeCount += zone.zoneQuestArchetypeCount;
          totals.challengeCount += zone.challengeCount;
          totals.scenarioCount += zone.scenarioCount;
          totals.monsterCount += zone.monsterCount;
          totals.monsterEncounterCount += zone.monsterEncounterCount;
          totals.treasureChestCount += zone.treasureChestCount;
          totals.healingFountainCount += zone.healingFountainCount;
          return totals;
        },
        {
          pointOfInterestCount: 0,
          questCount: 0,
          zoneQuestArchetypeCount: 0,
          challengeCount: 0,
          scenarioCount: 0,
          monsterCount: 0,
          monsterEncounterCount: 0,
          treasureChestCount: 0,
          healingFountainCount: 0,
        }
      ),
    [zoneSummaries]
  );

  const filteredZoneSummaries = useMemo(() => {
    const query = zoneSearchQuery.trim().toLowerCase();
    const filtered = zoneSummaries.filter((zone) => {
      if (!query) {
        return true;
      }
      return buildZoneSearchText(zone).includes(query);
    });

    return filtered.sort((left, right) => {
      if (zoneSort === 'name') {
        return left.name.localeCompare(right.name);
      }

      if (zoneSort === 'updated') {
        return (
          new Date(right.updatedAt).getTime() - new Date(left.updatedAt).getTime()
        );
      }

      if (zoneSort === 'quests') {
        return (
          right.questCount - left.questCount ||
          right.zoneQuestArchetypeCount - left.zoneQuestArchetypeCount ||
          left.name.localeCompare(right.name)
        );
      }

      const leftRichness =
        left.pointOfInterestCount +
        left.questCount +
        left.zoneQuestArchetypeCount +
        left.challengeCount +
        left.scenarioCount +
        left.monsterCount +
        left.monsterEncounterCount +
        left.treasureChestCount +
        left.healingFountainCount;
      const rightRichness =
        right.pointOfInterestCount +
        right.questCount +
        right.zoneQuestArchetypeCount +
        right.challengeCount +
        right.scenarioCount +
        right.monsterCount +
        right.monsterEncounterCount +
        right.treasureChestCount +
        right.healingFountainCount;

      return rightRichness - leftRichness || left.name.localeCompare(right.name);
    });
  }, [zoneSearchQuery, zoneSort, zoneSummaries]);

  const openCreateZoneModal = () => {
    setCreateMapCenter(
      mostRecentlyCreatedZone
        ? [mostRecentlyCreatedZone.longitude, mostRecentlyCreatedZone.latitude]
        : [-87.6298, 41.8781]
    );
    setName('');
    setDescription('');
    setInternalTagsInput('');
    setCreateBoundaryPoints([]);
    setCreateZoneError(null);
    setShowCreateZone(true);
  };

  const closeCreateZoneModal = () => {
    setShowCreateZone(false);
    setCreateZoneError(null);
    setIsCreatingZone(false);
  };

  const metroOptions = [
    'Atlanta, Georgia',
    'Austin, Texas',
    'Boston, Massachusetts',
    'Chicago, Illinois',
    'Dallas, Texas',
    'Denver, Colorado',
    'Houston, Texas',
    'Los Angeles, California',
    'Miami, Florida',
    'New York City, New York',
    'Philadelphia, Pennsylvania',
    'Phoenix, Arizona',
    'San Diego, California',
    'San Francisco, California',
    'Seattle, Washington',
    'Washington, DC',
  ];

  const effectiveMetro =
    selectedMetro === '__custom__' ? customMetro.trim() : selectedMetro;

  const handleImportZones = async () => {
    setImportError(null);
    if (!effectiveMetro) {
      setImportError('Please select a metro area.');
      return;
    }
    setImporting(true);
    try {
      const importItem = await apiClient.post<ZoneImport>(
        '/sonar/zones/import',
        {
          metroName: effectiveMetro,
        }
      );
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Error importing zones:', error);
      setImportError('Failed to start zone import.');
    } finally {
      setImporting(false);
    }
  };

  const fetchImportJobs = useCallback(async () => {
    try {
      const query = effectiveMetro
        ? `?metroName=${encodeURIComponent(effectiveMetro)}`
        : '';
      const response = await apiClient.get<ZoneImport[]>(
        `/sonar/zones/imports${query}`
      );
      setImportJobs(response);
      const hasPending = response.some(
        (item) => item.status === 'queued' || item.status === 'in_progress'
      );
      setImportPolling(hasPending);
    } catch (error) {
      console.error('Failed to fetch zone import status', error);
    }
  }, [apiClient, effectiveMetro]);

  const handleDeleteImportZones = async (importId: string) => {
    const confirmed = window.confirm(
      'Delete all zones created by this import? This cannot be undone.'
    );
    if (!confirmed) {
      return;
    }
    setImportError(null);
    setDeletingImportId(importId);
    try {
      await apiClient.delete(`/sonar/zones/imports/${importId}`);
      await fetchImportJobs();
      await refreshZones();
      await fetchZoneSummaries();
    } catch (error) {
      console.error('Failed to delete imported zones', error);
      setImportError('Failed to delete imported zones.');
    } finally {
      setDeletingImportId(null);
    }
  };

  const handleCreateZone = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setCreateZoneError('Zone name is required.');
      return;
    }

    if (createBoundaryPoints.length < 3) {
      setCreateZoneError('Please add at least 3 boundary points.');
      return;
    }

    const sortedBoundaryPoints = sortBoundaryPoints(createBoundaryPoints);
    const center = calculateBoundaryCenter(sortedBoundaryPoints);
    if (!center) {
      setCreateZoneError(
        'Unable to calculate the zone center from the boundary.'
      );
      return;
    }

    setCreateZoneError(null);
    setIsCreatingZone(true);

    try {
      const createdZone = await apiClient.post<Zone>('/sonar/zones', {
        name: trimmedName,
        description,
        internalTags: parseInternalTagsInput(internalTagsInput),
        latitude: center.latitude,
        longitude: center.longitude,
      });

      await apiClient.post(`/sonar/zones/${createdZone.id}/boundary`, {
        boundary: sortedBoundaryPoints.slice(0, -1),
      });

      await refreshZones();
      await fetchZoneSummaries();
      closeCreateZoneModal();
      navigate(`/zones/${createdZone.id}`);
    } catch (error) {
      console.error('Error creating zone:', error);
      setCreateZoneError('Failed to create zone.');
    } finally {
      setIsCreatingZone(false);
    }
  };

  useEffect(() => {
    void fetchImportJobs();
  }, [fetchImportJobs]);

  useEffect(() => {
    void fetchZoneSummaries();
  }, [fetchZoneSummaries]);

  useEffect(() => {
    if (!importPolling) return;
    const interval = setInterval(() => {
      void fetchImportJobs();
    }, 3000);
    return () => clearInterval(interval);
  }, [fetchImportJobs, importPolling]);

  useEffect(() => {
    if (importJobs.length === 0) return;
    const completed = importJobs.filter(
      (job) => job.status === 'completed' && job.zoneCount > 0
    );
    if (completed.length === 0) return;

    setNotifiedImportIds((prev) => {
      const next = new Set(prev);
      let hasNew = false;
      completed.forEach((job) => {
        if (!next.has(job.id)) {
          next.add(job.id);
          hasNew = true;
        }
      });
      if (hasNew) {
        void refreshZones();
        void fetchZoneSummaries();
      }
      return next;
    });
  }, [fetchZoneSummaries, importJobs, refreshZones]);

  const handleDeleteZone = async (zoneID: string) => {
    const confirmed = window.confirm(
      'Delete this zone and its zone-level content? This cannot be undone.'
    );
    if (!confirmed) {
      return;
    }

    setDeletingZoneId(zoneID);
    try {
      await apiClient.delete(`/sonar/zones/${zoneID}`);
      await refreshZones();
      await fetchZoneSummaries();
    } catch (error) {
      console.error('Error deleting zone:', error);
      setZoneSummariesError('Failed to delete zone.');
    } finally {
      setDeletingZoneId(null);
    }
  };

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

      mapRef.current.on('load', () => {
        setMapLoaded(true);
      });
    }

    return () => {
      popupRef.current?.remove();
      popupRef.current = null;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
      handlersAttachedRef.current = false;
      fitBoundsRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (!mapRef.current || !mapLoaded) {
      return;
    }

    const map = mapRef.current;
    const features = zones
      .map((zone) => {
        const rawCoords = getZoneRing(zone);
        const summary = zoneSummaryByID.get(zone.id);

        if (rawCoords.length < 4) {
          return null;
        }

        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: [rawCoords],
          },
          properties: {
            id: zone.id,
            name: zone.name,
            description: zone.description || '',
            boundaryCount: Math.max(rawCoords.length - 1, 0),
            pointOfInterestCount: summary?.pointOfInterestCount ?? 0,
            questCount: summary?.questCount ?? 0,
            zoneQuestArchetypeCount: summary?.zoneQuestArchetypeCount ?? 0,
            challengeCount: summary?.challengeCount ?? 0,
            scenarioCount: summary?.scenarioCount ?? 0,
            monsterEncounterCount: summary?.monsterEncounterCount ?? 0,
            importMetroName: summary?.importMetroName ?? '',
          },
        };
      })
      .filter(Boolean);

    const geojson = {
      type: 'FeatureCollection' as const,
      features: features as Array<GeoJSON.Feature<GeoJSON.Polygon>>,
    };

    const existingSource = map.getSource('zones') as
      | mapboxgl.GeoJSONSource
      | undefined;
    if (existingSource) {
      existingSource.setData(geojson);
    } else {
      map.addSource('zones', {
        type: 'geojson',
        data: geojson,
      });

      map.addLayer({
        id: 'zones-fill',
        type: 'fill',
        source: 'zones',
        paint: {
          'fill-color': '#3b82f6',
          'fill-opacity': 0.25,
        },
      });

      map.addLayer({
        id: 'zones-outline',
        type: 'line',
        source: 'zones',
        paint: {
          'line-color': '#1d4ed8',
          'line-width': 2,
        },
      });
    }

    if (features.length > 0 && !fitBoundsRef.current) {
      const bounds = new mapboxgl.LngLatBounds();
      features.forEach((feature) => {
        feature.geometry.coordinates[0].forEach((coord) => {
          bounds.extend(coord as [number, number]);
        });
      });
      map.fitBounds(bounds, { padding: 40, maxZoom: 12 });
      fitBoundsRef.current = true;
    }

    if (!handlersAttachedRef.current && map.getLayer('zones-fill')) {
      map.on('mouseenter', 'zones-fill', () => {
        map.getCanvas().style.cursor = 'pointer';
      });
      map.on('mouseleave', 'zones-fill', () => {
        map.getCanvas().style.cursor = '';
      });

      map.on('click', 'zones-fill', (event) => {
        const feature = event.features?.[0];
        if (!feature || !feature.properties) {
          return;
        }

        const zoneId = feature.properties.id as string;
        const zoneName = feature.properties.name as string;
        const zoneDescription = feature.properties.description as string;
        const boundaryCount = Number(feature.properties.boundaryCount ?? 0);
        const pointOfInterestCount = Number(
          feature.properties.pointOfInterestCount ?? 0
        );
        const questCount = Number(feature.properties.questCount ?? 0);
        const zoneQuestArchetypeCount = Number(
          feature.properties.zoneQuestArchetypeCount ?? 0
        );
        const challengeCount = Number(feature.properties.challengeCount ?? 0);
        const scenarioCount = Number(feature.properties.scenarioCount ?? 0);
        const monsterEncounterCount = Number(
          feature.properties.monsterEncounterCount ?? 0
        );
        const importMetroName = feature.properties.importMetroName as string;

        popupRef.current?.remove();

        const popupContent = document.createElement('div');
        popupContent.className = 'text-sm text-slate-700';

        const title = document.createElement('div');
        title.className = 'text-base font-semibold text-slate-800';
        title.textContent = zoneName;
        popupContent.appendChild(title);

        const description = document.createElement('div');
        description.className = 'mt-1 text-xs text-slate-600';
        description.textContent = zoneDescription || 'No description.';
        popupContent.appendChild(description);

        if (importMetroName) {
          const importMeta = document.createElement('div');
          importMeta.className = 'mt-2 text-xs font-medium text-indigo-600';
          importMeta.textContent = `Imported from ${importMetroName}`;
          popupContent.appendChild(importMeta);
        }

        const meta = document.createElement('div');
        meta.className = 'mt-2 text-xs text-slate-500';
        meta.textContent = `Boundary points: ${boundaryCount}`;
        popupContent.appendChild(meta);

        const counts = document.createElement('div');
        counts.className = 'mt-2 text-xs text-slate-500';
        counts.textContent = `POIs: ${pointOfInterestCount} · Quests: ${questCount} · Archetypes: ${zoneQuestArchetypeCount}`;
        popupContent.appendChild(counts);

        const objectiveCounts = document.createElement('div');
        objectiveCounts.className = 'mt-1 text-xs text-slate-500';
        objectiveCounts.textContent = `Challenges: ${challengeCount} · Scenarios: ${scenarioCount} · Encounters: ${monsterEncounterCount}`;
        popupContent.appendChild(objectiveCounts);

        const button = document.createElement('button');
        button.className =
          'mt-3 w-full rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-indigo-700';
        button.textContent = 'Open Zone';
        button.addEventListener('click', () => {
          popupRef.current?.remove();
          navigate(`/zones/${zoneId}`);
        });
        popupContent.appendChild(button);

        popupRef.current = new mapboxgl.Popup({ closeOnClick: true })
          .setLngLat(event.lngLat)
          .setDOMContent(popupContent)
          .addTo(map);
      });

      handlersAttachedRef.current = true;
    }
  }, [zones, navigate, mapLoaded, zoneSummaryByID]);

  const allZoneBoundaries = zones
    .map((zone) => getZoneRing(zone))
    .filter((points) => points.length >= 4);

  return (
    <div className="m-10">
      <div className="flex items-center justify-between gap-4">
        <h1 className="text-2xl font-bold">Zones</h1>
        <button
          className="rounded-md bg-blue-500 px-4 py-2 font-semibold text-white"
          onClick={openCreateZoneModal}
        >
          Create Zone
        </button>
      </div>
      <div className="mt-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="text-lg font-semibold text-slate-800">
          Import Neighborhood Zones
        </h2>
        <p className="text-sm text-slate-500">
          Select a metro area to import neighborhood polygons from OSM.
        </p>
        <div className="mt-4 flex flex-col gap-3 md:flex-row md:items-end">
          <div className="flex-1">
            <label className="mb-1 block text-sm font-medium text-slate-700">
              Metro Area
            </label>
            <select
              value={selectedMetro}
              onChange={(e) => setSelectedMetro(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            >
              {metroOptions.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
              <option value="__custom__">Custom...</option>
            </select>
          </div>
          {selectedMetro === '__custom__' && (
            <div className="flex-1">
              <label className="mb-1 block text-sm font-medium text-slate-700">
                Custom Metro
              </label>
              <input
                type="text"
                value={customMetro}
                onChange={(e) => setCustomMetro(e.target.value)}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                placeholder="e.g., Minneapolis, Minnesota"
              />
            </div>
          )}
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
            onClick={handleImportZones}
            disabled={importing || !effectiveMetro}
          >
            {importing ? 'Queueing...' : 'Import Zones'}
          </button>
        </div>
        {importError && (
          <p className="mt-2 text-sm text-red-600">{importError}</p>
        )}
        <div className="mt-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-slate-700">
              Recent Imports
            </h3>
            <button
              className="text-xs font-semibold text-slate-500 hover:text-slate-700"
              onClick={fetchImportJobs}
            >
              Refresh
            </button>
          </div>
          {importJobs.length === 0 ? (
            <p className="mt-2 text-sm text-slate-400">No imports yet.</p>
          ) : (
            <div className="mt-2 space-y-2">
              {importJobs.slice(0, 6).map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between rounded-md border border-slate-200 px-3 py-2 text-sm"
                >
                  <div>
                    <div className="font-medium text-slate-800">
                      {job.metroName}
                    </div>
                    <div className="text-xs text-slate-500">
                      Status: {job.status}
                    </div>
                    {job.errorMessage && (
                      <div className="text-xs text-red-600">
                        {job.errorMessage}
                      </div>
                    )}
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="text-xs text-slate-500">
                      Zones: {job.zoneCount}
                    </div>
                    <button
                      className="rounded-md border border-slate-200 px-2 py-1 text-xs font-semibold text-slate-600 hover:text-slate-800 disabled:opacity-60"
                      onClick={() => handleDeleteImportZones(job.id)}
                      disabled={deletingImportId === job.id}
                    >
                      {deletingImportId === job.id
                        ? 'Deleting...'
                        : 'Delete Zones'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
      <div className="mt-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-800">
              Zone Boundaries
            </h2>
            <p className="text-sm text-slate-500">
              Click a zone polygon to view details and open its page.
            </p>
          </div>
          <button
            className="text-xs font-semibold text-slate-500 hover:text-slate-700"
            onClick={() => {
              if (!mapRef.current || zones.length === 0) return;
              const bounds = new mapboxgl.LngLatBounds();
              let hasBounds = false;
              zones.forEach((zone) => {
                const coords = getZoneRing(zone);
                coords.forEach((coord) => {
                  bounds.extend(coord);
                  hasBounds = true;
                });
              });
              if (hasBounds) {
                mapRef.current.fitBounds(bounds, { padding: 40, maxZoom: 12 });
              }
            }}
          >
            Fit to zones
          </button>
        </div>
        <div className="mt-4 h-[420px] w-full overflow-hidden rounded-lg border border-slate-200">
          <div ref={mapContainer} className="h-full w-full" />
        </div>
      </div>
      <div className="mt-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-800">
              Zone Library
            </h2>
            <p className="text-sm text-slate-500">
              Browse each neighborhood with content totals, import provenance,
              internal tags, and quick health signals.
            </p>
          </div>
          <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {filteredZoneSummaries.length} showing of {zoneSummaries.length}{' '}
            zones
          </div>
        </div>

        <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Zones
            </div>
            <div className="mt-2 text-2xl font-semibold text-slate-900">
              {zoneSummaries.length}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              Neighborhoods loaded into admin
            </div>
          </div>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Points Of Interest
            </div>
            <div className="mt-2 text-2xl font-semibold text-slate-900">
              {zoneOverview.pointOfInterestCount}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              Mapped places tied to zones
            </div>
          </div>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Quests
            </div>
            <div className="mt-2 text-2xl font-semibold text-slate-900">
              {zoneOverview.questCount}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              Live quests across all zones
            </div>
          </div>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Objective Content
            </div>
            <div className="mt-2 text-2xl font-semibold text-slate-900">
              {zoneOverview.challengeCount +
                zoneOverview.scenarioCount +
                zoneOverview.monsterEncounterCount}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              Challenges, scenarios, and encounters
            </div>
          </div>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Assigned Archetypes
            </div>
            <div className="mt-2 text-2xl font-semibold text-slate-900">
              {zoneOverview.zoneQuestArchetypeCount}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              Zone-level quest generation assignments
            </div>
          </div>
        </div>

        <div className="mt-4 flex flex-col gap-3 lg:flex-row">
          <div className="flex-1">
            <label className="mb-1 block text-sm font-medium text-slate-700">
              Search Zones
            </label>
            <input
              type="text"
              value={zoneSearchQuery}
              onChange={(event) => setZoneSearchQuery(event.target.value)}
              placeholder="Search by name, description, import metro, or internal tags"
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            />
          </div>
          <div className="w-full lg:w-64">
            <label className="mb-1 block text-sm font-medium text-slate-700">
              Sort By
            </label>
            <select
              value={zoneSort}
              onChange={(event) =>
                setZoneSort(
                  event.target.value as 'richest' | 'quests' | 'updated' | 'name'
                )
              }
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            >
              <option value="richest">Most Content</option>
              <option value="quests">Most Quests</option>
              <option value="updated">Recently Updated</option>
              <option value="name">Alphabetical</option>
            </select>
          </div>
        </div>

        {zoneSummariesError && (
          <div className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {zoneSummariesError}
          </div>
        )}

        {zoneSummariesLoading ? (
          <div className="mt-6 text-sm text-slate-500">Loading zone summaries...</div>
        ) : filteredZoneSummaries.length === 0 ? (
          <div className="mt-6 rounded-lg border border-dashed border-slate-200 bg-slate-50 px-4 py-8 text-center text-sm text-slate-500">
            No zones matched that search.
          </div>
        ) : (
          <div className="mt-6 grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
            {filteredZoneSummaries.map((zone) => {
              const contentCounts = [
                { label: 'POIs', value: zone.pointOfInterestCount },
                { label: 'Quests', value: zone.questCount },
                {
                  label: 'Archetypes',
                  value: zone.zoneQuestArchetypeCount,
                },
                { label: 'Challenges', value: zone.challengeCount },
                { label: 'Scenarios', value: zone.scenarioCount },
                {
                  label: 'Encounters',
                  value: zone.monsterEncounterCount,
                },
                { label: 'Monsters', value: zone.monsterCount },
                { label: 'Chests', value: zone.treasureChestCount },
                {
                  label: 'Fountains',
                  value: zone.healingFountainCount,
                },
              ];

              return (
                <article
                  key={zone.id}
                  className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="text-lg font-semibold text-slate-900">
                        {zone.name}
                      </h3>
                      <div className="mt-1 text-xs font-medium uppercase tracking-wide text-slate-500">
                        {zone.importMetroName
                          ? `Imported from ${zone.importMetroName}`
                          : 'Manual zone'}
                      </div>
                    </div>
                    <div className="rounded-full bg-indigo-50 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-indigo-700">
                      {getZoneActivityLabel(zone)}
                    </div>
                  </div>

                  <p className="mt-3 min-h-[60px] text-sm leading-6 text-slate-600">
                    {zone.description || 'No description yet.'}
                  </p>

                  <div className="mt-4 flex flex-wrap gap-2">
                    {(zone.internalTags && zone.internalTags.length > 0
                      ? zone.internalTags
                      : ['untagged']
                    ).map((tag) => (
                      <span
                        key={`${zone.id}-${tag}`}
                        className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>

                  <div className="mt-4 grid gap-3 sm:grid-cols-2">
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                      <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                        Center
                      </div>
                      <div className="mt-1 text-sm font-medium text-slate-800">
                        {formatCoordinate(zone.latitude)},{' '}
                        {formatCoordinate(zone.longitude)}
                      </div>
                    </div>
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                      <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                        Boundary
                      </div>
                      <div className="mt-1 text-sm font-medium text-slate-800">
                        {zone.boundaryPointCount} point
                        {zone.boundaryPointCount === 1 ? '' : 's'}
                      </div>
                    </div>
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                      <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                        Updated
                      </div>
                      <div className="mt-1 text-sm font-medium text-slate-800">
                        {formatZoneDate(zone.updatedAt)}
                      </div>
                    </div>
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                      <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                        Created
                      </div>
                      <div className="mt-1 text-sm font-medium text-slate-800">
                        {formatZoneDate(zone.createdAt)}
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 grid grid-cols-3 gap-2">
                    {contentCounts.map((item) => (
                      <div
                        key={`${zone.id}-${item.label}`}
                        className="rounded-lg border border-slate-200 px-3 py-2"
                      >
                        <div className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                          {item.label}
                        </div>
                        <div className="mt-1 text-lg font-semibold text-slate-900">
                          {item.value}
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="mt-4 rounded-lg bg-indigo-50 p-3">
                    <div className="text-xs font-semibold uppercase tracking-wide text-indigo-700">
                      Zone Readiness
                    </div>
                    <p className="mt-1 text-sm leading-6 text-indigo-900">
                      {getZoneReadinessSummary(zone)}
                    </p>
                  </div>

                  <div className="mt-4 flex flex-wrap gap-2">
                    <button
                      onClick={() => navigate(`/zones/${zone.id}`)}
                      className="rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700"
                    >
                      View Zone
                    </button>
                    <button
                      onClick={() => void handleDeleteZone(zone.id)}
                      disabled={deletingZoneId === zone.id}
                      className="rounded-md border border-red-200 px-4 py-2 text-sm font-semibold text-red-700 hover:bg-red-50 disabled:opacity-60"
                    >
                      {deletingZoneId === zone.id ? 'Deleting...' : 'Delete Zone'}
                    </button>
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </div>
      {showCreateZone && (
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
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '20px',
              borderRadius: '8px',
              width: 'min(900px, calc(100vw - 40px))',
              maxHeight: 'calc(100vh - 40px)',
              overflowY: 'auto',
            }}
          >
            <h2>Create Zone</h2>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Name:
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: '4px',
                  border: '1px solid #ccc',
                }}
              />
            </div>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Description:
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: '4px',
                  border: '1px solid #ccc',
                  minHeight: '100px',
                  resize: 'vertical',
                }}
              />
            </div>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Internal Tags:
              </label>
              <input
                type="text"
                value={internalTagsInput}
                onChange={(e) => setInternalTagsInput(e.target.value)}
                placeholder="e.g. starter_region, downtown, high_density"
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: '4px',
                  border: '1px solid #ccc',
                }}
              />
              <div
                style={{ marginTop: '6px', color: '#666', fontSize: '12px' }}
              >
                Internal-only metadata tags. These are not shown to players.
              </div>
            </div>
            <ZoneBoundaryEditorMap
              center={createMapCenter}
              boundaryPoints={createBoundaryPoints}
              allZoneBoundaries={allZoneBoundaries}
              onMapClick={(lngLat) => {
                setCreateBoundaryPoints((prev) => [
                  ...prev,
                  [lngLat.lng, lngLat.lat],
                ]);
              }}
              onClearBoundary={() => setCreateBoundaryPoints([])}
            />
            {createZoneError && (
              <div
                style={{
                  marginBottom: '15px',
                  color: '#dc2626',
                  fontSize: '14px',
                }}
              >
                {createZoneError}
              </div>
            )}
            <div
              style={{
                display: 'flex',
                justifyContent: 'flex-end',
                gap: '10px',
              }}
            >
              <button
                onClick={closeCreateZoneModal}
                style={{
                  padding: '8px 16px',
                  borderRadius: '4px',
                  border: '1px solid #ccc',
                  backgroundColor: '#fff',
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleCreateZone}
                style={{
                  padding: '8px 16px',
                  borderRadius: '4px',
                  border: 'none',
                  backgroundColor: '#007bff',
                  color: '#fff',
                }}
                disabled={isCreatingZone}
              >
                {isCreatingZone ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
