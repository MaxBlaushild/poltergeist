import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useZoneContext } from '@poltergeist/contexts';
import { v4 as uuidv4 } from 'uuid';
import { useNavigate, useParams } from 'react-router-dom';
import { useZonePointsOfInterest } from '../hooks/useZonePointsOfInterest.ts';
import { usePlaceTypes } from '../hooks/usePlaceTypes.ts';
import { useGeneratePointsOfInterest } from '../hooks/useGeneratePointsOfInterest.ts';
import { useCandidates } from '@poltergeist/hooks';
import {
  Candidate,
  Character,
  CharacterLocation,
  TreasureChest,
} from '@poltergeist/types';
import { useQuestArchtypes } from '../hooks/useQuestArchtypes.ts';
import { useAPI } from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { Buffer } from 'buffer';
import * as turf from '@turf/turf';
import * as wellknown from 'wellknown';
import { Geometry, Polygon } from 'wkx-ts';
import wkx from 'wkx';
// Set Mapbox access token
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

const normalizeZoneRouteKey = (value: string): string =>
  value
    .trim()
    .toLowerCase()
    .replace(/[_-]+/g, ' ')
    .split(/\s+/)
    .filter(Boolean)
    .join(' ');

interface MapProps {
  center: [number, number];
  onMapClick?: (lngLat: mapboxgl.LngLat) => void;
  boundaryPoints?: [number, number][];
  allZoneBoundaries?: [number, number][][];
  showBoundaryControls?: boolean;
  pins?: AdminMapPin[];
  pinDragEnabled?: boolean;
  onPinLocationChange?: (
    pin: AdminMapPin,
    coordinates: [number, number]
  ) => Promise<void> | void;
}

type HealingFountainRecord = {
  id: string;
  name: string;
  thumbnailUrl: string;
  latitude: number;
  longitude: number;
};

type ScenarioRecord = {
  id: string;
  pointOfInterestId?: string | null;
  prompt: string;
  latitude: number;
  longitude: number;
  imageUrl: string;
  thumbnailUrl: string;
};

type MonsterEncounterRecord = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  imageUrl: string;
  thumbnailUrl: string;
};

type ChallengeRecord = {
  id: string;
  pointOfInterestId?: string | null;
  question: string;
  latitude: number;
  longitude: number;
  imageUrl: string;
  thumbnailUrl: string;
  polygonPoints?: [number, number][] | null;
};

type ExpositionRecord = {
  id: string;
  pointOfInterestId?: string | null;
  title: string;
  latitude: number;
  longitude: number;
  imageUrl: string;
  thumbnailUrl: string;
};

type AdminMapPinKind =
  | 'pointOfInterest'
  | 'character'
  | 'treasureChest'
  | 'healingFountain'
  | 'scenario'
  | 'exposition'
  | 'monster'
  | 'challenge';

type AdminMapPin = {
  id: string;
  entityId: string;
  kind: AdminMapPinKind;
  name: string;
  coordinates: [number, number];
  imageUrl: string;
  draggable: boolean;
  locationIndex?: number;
  dragHint?: string;
};

const pinDeleteLabelByKind: Record<AdminMapPinKind, string> = {
  pointOfInterest: 'point of interest',
  character: 'character',
  treasureChest: 'treasure chest',
  healingFountain: 'healing fountain',
  scenario: 'scenario',
  exposition: 'exposition',
  monster: 'monster encounter',
  challenge: 'challenge',
};

const chestImageUrl =
  'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';
const poiMysteryImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const characterMysteryImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png';
const scenarioMysteryImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png';
const expositionMysteryImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/exposition-undiscovered.png';
const monsterMysteryImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png';
const healingFountainDiscoveredImageUrl =
  'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/healing-fountain-discovered.png';

const markerStyleByKind: Record<
  AdminMapPinKind,
  {
    ring: string;
    badge: string;
    shortLabel: string;
    fullLabel: string;
    fallbackImage: string;
  }
> = {
  pointOfInterest: {
    ring: '#2563eb',
    badge: '#1d4ed8',
    shortLabel: 'POI',
    fullLabel: 'Point of Interest',
    fallbackImage: poiMysteryImageUrl,
  },
  character: {
    ring: '#ec4899',
    badge: '#be185d',
    shortLabel: 'NPC',
    fullLabel: 'Character',
    fallbackImage: characterMysteryImageUrl,
  },
  treasureChest: {
    ring: '#f59e0b',
    badge: '#b45309',
    shortLabel: 'TC',
    fullLabel: 'Treasure Chest',
    fallbackImage: chestImageUrl,
  },
  healingFountain: {
    ring: '#10b981',
    badge: '#047857',
    shortLabel: 'HF',
    fullLabel: 'Healing Fountain',
    fallbackImage: healingFountainDiscoveredImageUrl,
  },
  scenario: {
    ring: '#14b8a6',
    badge: '#0f766e',
    shortLabel: 'SC',
    fullLabel: 'Scenario',
    fallbackImage: scenarioMysteryImageUrl,
  },
  exposition: {
    ring: '#f59e0b',
    badge: '#b45309',
    shortLabel: 'EX',
    fullLabel: 'Exposition',
    fallbackImage: expositionMysteryImageUrl,
  },
  monster: {
    ring: '#ef4444',
    badge: '#991b1b',
    shortLabel: 'MO',
    fullLabel: 'Monster',
    fallbackImage: monsterMysteryImageUrl,
  },
  challenge: {
    ring: '#8b5cf6',
    badge: '#6d28d9',
    shortLabel: 'CH',
    fullLabel: 'Challenge',
    fallbackImage: scenarioMysteryImageUrl,
  },
};

const isValidCoordinate = (latitude: number, longitude: number) =>
  Number.isFinite(latitude) &&
  Number.isFinite(longitude) &&
  !(latitude === 0 && longitude === 0);

const createPinPopup = (pin: AdminMapPin) => {
  const style = markerStyleByKind[pin.kind];
  const container = document.createElement('div');
  container.style.display = 'flex';
  container.style.flexDirection = 'column';
  container.style.gap = '2px';

  const title = document.createElement('div');
  title.style.fontWeight = '600';
  title.style.color = '#0f172a';
  title.textContent = pin.name;
  container.appendChild(title);

  const subtitle = document.createElement('div');
  subtitle.style.fontSize = '12px';
  subtitle.style.color = '#475569';
  subtitle.textContent = style.fullLabel;
  container.appendChild(subtitle);

  return container;
};

const createPinElement = (
  pin: AdminMapPin,
  interactive: boolean,
  draggable = false
) => {
  const style = markerStyleByKind[pin.kind];
  const root = document.createElement('button');
  root.type = 'button';
  root.title = `${style.fullLabel}: ${pin.name}`;
  root.style.width = '48px';
  root.style.height = '60px';
  root.style.display = 'flex';
  root.style.flexDirection = 'column';
  root.style.alignItems = 'center';
  root.style.justifyContent = 'flex-start';
  root.style.padding = '0';
  root.style.background = 'transparent';
  root.style.border = 'none';
  root.style.cursor = draggable ? 'grab' : interactive ? 'pointer' : 'default';
  root.style.pointerEvents = interactive || draggable ? 'auto' : 'none';

  const imageFrame = document.createElement('div');
  imageFrame.style.width = '40px';
  imageFrame.style.height = '40px';
  imageFrame.style.borderRadius = '9999px';
  imageFrame.style.overflow = 'hidden';
  imageFrame.style.border = `3px solid ${style.ring}`;
  imageFrame.style.background = '#e2e8f0';
  imageFrame.style.boxShadow = '0 8px 18px rgba(15, 23, 42, 0.28)';

  const image = document.createElement('img');
  image.src = pin.imageUrl || style.fallbackImage;
  image.alt = pin.name;
  image.width = 40;
  image.height = 40;
  image.style.width = '100%';
  image.style.height = '100%';
  image.style.objectFit = 'cover';
  image.loading = 'lazy';
  image.onerror = () => {
    if (image.src !== style.fallbackImage) {
      image.src = style.fallbackImage;
    }
  };
  imageFrame.appendChild(image);

  const badge = document.createElement('div');
  badge.textContent = style.shortLabel;
  badge.style.marginTop = '-8px';
  badge.style.padding = '2px 6px';
  badge.style.borderRadius = '9999px';
  badge.style.background = style.badge;
  badge.style.color = '#ffffff';
  badge.style.fontSize = '9px';
  badge.style.fontWeight = '700';
  badge.style.letterSpacing = '0.06em';
  badge.style.boxShadow = '0 4px 12px rgba(15, 23, 42, 0.25)';

  root.appendChild(imageFrame);
  root.appendChild(badge);

  return root;
};

const sortBoundaryPoints = (points: [number, number][]) => {
  const sortedPoints = points.slice();
  if (sortedPoints.length === 0) {
    return sortedPoints;
  }

  const centroid = sortedPoints.reduce(
    (acc, point) => {
      return [
        acc[0] + point[0] / sortedPoints.length,
        acc[1] + point[1] / sortedPoints.length,
      ];
    },
    [0, 0]
  );

  sortedPoints.sort((a, b) => {
    const angleA = Math.atan2(a[1] - centroid[1], a[0] - centroid[0]);
    const angleB = Math.atan2(b[1] - centroid[1], b[0] - centroid[0]);
    return angleA - angleB;
  });

  if (
    sortedPoints.length > 0 &&
    (sortedPoints[0][0] !== sortedPoints[sortedPoints.length - 1][0] ||
      sortedPoints[0][1] !== sortedPoints[sortedPoints.length - 1][1])
  ) {
    sortedPoints.push([sortedPoints[0][0], sortedPoints[0][1]]);
  }

  return sortedPoints;
};

const ZoneMap: React.FC<MapProps> = ({
  center,
  onMapClick,
  boundaryPoints,
  allZoneBoundaries,
  showBoundaryControls,
  pins = [],
  pinDragEnabled = false,
  onPinLocationChange,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const boundaryMarkers = useRef<mapboxgl.Marker[]>([]);
  const pinMarkers = useRef<mapboxgl.Marker[]>([]);
  const [isLocating, setIsLocating] = useState(false);
  const [locationError, setLocationError] = useState<string | null>(null);

  // Clean up function to remove sources and layers
  const cleanupBoundary = () => {
    if (map.current) {
      if (map.current.getLayer('zone-boundary-outline')) {
        map.current.removeLayer('zone-boundary-outline');
      }
      if (map.current.getLayer('zone-boundary')) {
        map.current.removeLayer('zone-boundary');
      }
      if (map.current.getSource('zone-boundary')) {
        map.current.removeSource('zone-boundary');
      }
    }
  };

  useEffect(() => {
    if (mapContainer.current && !map.current) {
      map.current = new mapboxgl.Map({
        container: mapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: center,
        zoom: 14,
        interactive: true,
      });

      map.current.on('load', () => {
        setMapLoaded(true);
      });
    }

    return () => {
      cleanupBoundary();
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (map.current && mapLoaded && onMapClick) {
      map.current.dragPan.disable();
      const canvasContainer = mapContainer.current?.querySelector(
        '.mapboxgl-canvas-container'
      );
      if (canvasContainer instanceof HTMLElement) {
        canvasContainer.style.cursor = 'default';
      }
    } else if (map.current && mapLoaded) {
      map.current.dragPan.enable();
      const canvasContainer = mapContainer.current?.querySelector(
        '.mapboxgl-canvas-container'
      );
      if (canvasContainer instanceof HTMLElement) {
        canvasContainer.style.cursor = 'grab';
      }
    }

    if (!map.current || !mapLoaded || !onMapClick) {
      return;
    }

    const handleClick = (e: mapboxgl.MapMouseEvent) => {
      onMapClick(e.lngLat);
    };

    map.current.on('click', handleClick);
    return () => {
      map.current?.off('click', handleClick);
    };
  }, [mapLoaded, onMapClick]);

  useEffect(() => {
    if (map.current && mapLoaded) {
      // map.current.setCenter(center);
    }
  }, [center, mapLoaded]);

  // Update markers when boundary points change
  useEffect(() => {
    if (map.current && mapLoaded) {
      // Remove existing markers
      boundaryMarkers.current.forEach((marker) => marker.remove());
      boundaryMarkers.current = [];

      const sortedPoints = boundaryPoints
        ? sortBoundaryPoints(boundaryPoints)
        : [];
      if (sortedPoints.length > 0) {
        // Add markers for the sorted points
        sortedPoints.forEach((point) => {
          const marker = new mapboxgl.Marker({
            color: '#DC2626',
            draggable: false,
          })
            .setLngLat(point)
            .addTo(map.current!);
          boundaryMarkers.current.push(marker);
        });
      }
    }
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (!map.current || !mapLoaded) {
      return;
    }

    pinMarkers.current.forEach((marker) => marker.remove());
    pinMarkers.current = [];

    pins.forEach((pin) => {
      const marker = new mapboxgl.Marker({
        element: createPinElement(
          pin,
          !onMapClick || pinDragEnabled,
          pinDragEnabled && pin.draggable
        ),
        anchor: 'bottom',
        draggable: pinDragEnabled && pin.draggable,
      }).setLngLat(pin.coordinates);

      let lastSavedCoordinates = pin.coordinates;
      if (!pinDragEnabled) {
        marker.setPopup(
          new mapboxgl.Popup({
            offset: 18,
            closeButton: false,
            className: 'zone-map-pin-popup',
          }).setDOMContent(createPinPopup(pin))
        );
      }

      if (pinDragEnabled && pin.draggable && onPinLocationChange) {
        marker.on('dragend', async () => {
          const lngLat = marker.getLngLat();
          const nextCoordinates: [number, number] = [lngLat.lng, lngLat.lat];
          try {
            await onPinLocationChange(pin, nextCoordinates);
            lastSavedCoordinates = nextCoordinates;
          } catch (_) {
            marker.setLngLat(lastSavedCoordinates);
          }
        });
      }

      marker.addTo(map.current!);

      pinMarkers.current.push(marker);
    });

    return () => {
      pinMarkers.current.forEach((marker) => marker.remove());
      pinMarkers.current = [];
    };
  }, [mapLoaded, onMapClick, onPinLocationChange, pinDragEnabled, pins]);

  useEffect(() => {
    if (
      map.current &&
      mapLoaded &&
      boundaryPoints &&
      boundaryPoints.length > 0
    ) {
      // Clean up existing boundary
      cleanupBoundary();

      const sortedPoints = sortBoundaryPoints(boundaryPoints);
      if (sortedPoints.length > 0) {
        // Add boundary polygon
        map.current.addSource('zone-boundary', {
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
          id: 'zone-boundary',
          type: 'fill',
          source: 'zone-boundary',
          layout: {},
          paint: {
            'fill-color': '#DC2626',
            'fill-opacity': 0.3,
          },
        });

        // Add boundary outline
        map.current.addLayer({
          id: 'zone-boundary-outline',
          type: 'line',
          source: 'zone-boundary',
          layout: {},
          paint: {
            'line-color': '#DC2626',
            'line-width': 2,
          },
        });
      }
    } else if (map.current && mapLoaded) {
      // Clean up if no boundary points
      cleanupBoundary();
    }
  }, [boundaryPoints, mapLoaded]);

  useEffect(() => {
    if (map.current && mapLoaded) {
      if (map.current.getLayer('all-zone-boundaries-outline')) {
        map.current.removeLayer('all-zone-boundaries-outline');
      }
      if (map.current.getLayer('all-zone-boundaries')) {
        map.current.removeLayer('all-zone-boundaries');
      }
      if (map.current.getSource('all-zone-boundaries')) {
        map.current.removeSource('all-zone-boundaries');
      }
    }

    const polygonFeatures = (allZoneBoundaries || [])
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

    if (map.current && mapLoaded && polygonFeatures.length > 0) {
      map.current.addSource('all-zone-boundaries', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: polygonFeatures,
        },
      });

      map.current.addLayer(
        {
          id: 'all-zone-boundaries',
          type: 'fill',
          source: 'all-zone-boundaries',
          layout: {},
          paint: {
            'fill-color': '#2563EB',
            'fill-opacity': 0.18,
          },
        },
        map.current.getLayer('zone-boundary') ? 'zone-boundary' : undefined
      );

      map.current.addLayer(
        {
          id: 'all-zone-boundaries-outline',
          type: 'line',
          source: 'all-zone-boundaries',
          layout: {},
          paint: {
            'line-color': '#2563EB',
            'line-width': 2,
          },
        },
        map.current.getLayer('zone-boundary-outline')
          ? 'zone-boundary-outline'
          : undefined
      );
    }

    return () => {
      if (map.current) {
        if (map.current.getLayer('all-zone-boundaries-outline')) {
          map.current.removeLayer('all-zone-boundaries-outline');
        }
        if (map.current.getLayer('all-zone-boundaries')) {
          map.current.removeLayer('all-zone-boundaries');
        }
        if (map.current.getSource('all-zone-boundaries')) {
          map.current.removeSource('all-zone-boundaries');
        }
      }
    };
  }, [allZoneBoundaries, mapLoaded]);

  return (
    <div className="relative w-full h-96 rounded-lg border border-gray-300 overflow-hidden">
      <div ref={mapContainer} className="w-full h-full" />
      {showBoundaryControls && (
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
                  map.current?.flyTo({
                    center: [longitude, latitude],
                    zoom: Math.max(map.current?.getZoom() ?? 14, 16),
                    essential: true,
                  });
                },
                (err) => {
                  setIsLocating(false);
                  setLocationError(err.message || 'Unable to fetch location.');
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
      )}
    </div>
  );
};

type ZoneFlavorGenerationJob = {
  id: string;
  zoneId: string;
  status: string;
  generatedDescription?: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
};

const isZoneFlavorPendingStatus = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

type ZoneTagGenerationJob = {
  id: string;
  zoneId: string;
  status: string;
  contextSnapshot?: string;
  generatedSummary?: string;
  selectedTags?: string[];
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
};

const isZoneTagPendingStatus = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

export const Zone = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { apiClient } = useAPI();
  const { zones, createZone, deleteZone, refreshZones } = useZoneContext();
  const zone = useMemo(() => {
    if (!id) {
      return undefined;
    }
    const directMatch = zones.find((entry) => entry.id === id);
    if (directMatch) {
      return directMatch;
    }
    const normalizedRouteKey = normalizeZoneRouteKey(id);
    return zones.find(
      (entry) => normalizeZoneRouteKey(entry.name) === normalizedRouteKey
    );
  }, [id, zones]);
  const resolvedZoneId = zone?.id ?? id ?? '';
  const { pointsOfInterest, loading, error } =
    useZonePointsOfInterest(resolvedZoneId);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [treasureChests, setTreasureChests] = useState<TreasureChest[]>([]);
  const [healingFountains, setHealingFountains] = useState<
    HealingFountainRecord[]
  >([]);
  const [scenarios, setScenarios] = useState<ScenarioRecord[]>([]);
  const [expositions, setExpositions] = useState<ExpositionRecord[]>([]);
  const [monsterEncounters, setMonsterEncounters] = useState<
    MonsterEncounterRecord[]
  >([]);
  const [challenges, setChallenges] = useState<ChallengeRecord[]>([]);
  const [zoneMapPinsError, setZoneMapPinsError] = useState<string | null>(null);
  const {
    placeTypes,
    loading: placeTypesLoading,
    error: placeTypesError,
  } = usePlaceTypes();
  const [isGenerating, setIsGenerating] = useState(false);
  const [selectedIncludedPlaceTypes, setSelectedIncludedPlaceTypes] = useState<
    string[]
  >([]);
  const [selectedExcludedPlaceTypes, setSelectedExcludedPlaceTypes] = useState<
    string[]
  >([]);
  const {
    questArchtypes,
    loading: questArchtypesLoading,
    error: questArchtypesError,
  } = useQuestArchtypes();
  const [numPlaces, setNumPlaces] = useState(1);
  const [address, setAddress] = useState('');
  const [showPlaces, setShowPlaces] = useState(false);
  const timeoutRef = React.useRef<number>();
  const [query, setQuery] = useState('');
  const [shouldShowImage, setShouldShowImage] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [importedPlaces, setImportedPlaces] = useState<string[]>([]);
  const [isGeneratingQuest, setIsGeneratingQuest] = useState(false);
  const [selectedQuestArchtype, setSelectedQuestArchtype] = useState<
    string | null
  >(null);
  const [nameFilter, setNameFilter] = useState('');
  const [boundaryPoints, setBoundaryPoints] = useState<[number, number][]>([]);
  const [isEditingBoundary, setIsEditingBoundary] = useState(false);
  const [isEditingPins, setIsEditingPins] = useState(false);
  const [isEditingZone, setIsEditingZone] = useState(false);
  const [isSavingZone, setIsSavingZone] = useState(false);
  const [name, setName] = useState(zone?.name || '');
  const [description, setDescription] = useState(zone?.description || '');
  const [internalTagsInput, setInternalTagsInput] = useState(
    (zone?.internalTags ?? []).join(', ')
  );
  const [movingPinId, setMovingPinId] = useState<string | null>(null);
  const [pinMoveStatus, setPinMoveStatus] = useState<string | null>(null);
  const [pinMoveError, setPinMoveError] = useState<string | null>(null);
  const [deletingPinId, setDeletingPinId] = useState<string | null>(null);
  const [pinDeleteStatus, setPinDeleteStatus] = useState<string | null>(null);
  const [pinDeleteError, setPinDeleteError] = useState<string | null>(null);
  const [pinLocationOverrides, setPinLocationOverrides] = useState<
    Record<string, [number, number]>
  >({});
  const [deletedPointOfInterestIds, setDeletedPointOfInterestIds] = useState<
    Set<string>
  >(new Set());
  const [queueingZoneFlavor, setQueueingZoneFlavor] = useState(false);
  const [zoneFlavorJobs, setZoneFlavorJobs] = useState<
    ZoneFlavorGenerationJob[]
  >([]);
  const [zoneFlavorJobsError, setZoneFlavorJobsError] = useState<string | null>(
    null
  );
  const lastCompletedZoneFlavorJobIdRef = useRef<string | null>(null);
  const [queueingZoneTags, setQueueingZoneTags] = useState(false);
  const [zoneTagJobs, setZoneTagJobs] = useState<ZoneTagGenerationJob[]>([]);
  const [zoneTagJobsError, setZoneTagJobsError] = useState<string | null>(null);
  const lastCompletedZoneTagJobIdRef = useRef<string | null>(null);
  const {
    candidates,
    loading: candidatesLoading,
    error: candidatesError,
  } = useCandidates(query);

  const {
    loading: generatePointsOfInterestLoading,
    error: generatePointsOfInterestError,
    generatePointsOfInterest,
    refreshPointOfInterestImage,
    refreshPointOfInterest,
    importPointOfInterest,
    generateQuest,
  } = useGeneratePointsOfInterest(resolvedZoneId);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  useEffect(() => {
    if (!zone || !id || zone.id === id) {
      return;
    }
    navigate(`/zones/${zone.id}`, { replace: true });
  }, [id, navigate, zone]);

  useEffect(() => {
    if (zone?.points) {
      setBoundaryPoints(
        zone.points.map(
          (point) => [point.longitude, point.latitude] as [number, number]
        )
      );
    }
  }, [zone]);

  useEffect(() => {
    setName(zone?.name || '');
    setDescription(zone?.description || '');
    setInternalTagsInput((zone?.internalTags ?? []).join(', '));
  }, [zone?.name, zone?.description, zone?.internalTags]);

  useEffect(() => {
    setIsEditingBoundary(false);
    setIsEditingPins(false);
    setPinMoveStatus(null);
    setPinMoveError(null);
    setPinDeleteStatus(null);
    setPinDeleteError(null);
    setDeletingPinId(null);
    setPinLocationOverrides({});
    setDeletedPointOfInterestIds(new Set());
  }, [resolvedZoneId]);

  useEffect(() => {
    if (!resolvedZoneId) return;

    let active = true;
    const loadZoneMapPins = async () => {
      try {
        const [
          fetchedCharacters,
          fetchedTreasureChests,
          fetchedHealingFountains,
          fetchedScenarios,
          fetchedExpositions,
          fetchedMonsterEncounters,
          fetchedChallenges,
        ] = await Promise.all([
          apiClient.get<Character[]>('/sonar/characters'),
          apiClient.get<TreasureChest[]>(
            `/sonar/zones/${resolvedZoneId}/treasure-chests`
          ),
          apiClient.get<HealingFountainRecord[]>(
            `/sonar/zones/${resolvedZoneId}/healing-fountains`
          ),
          apiClient.get<ScenarioRecord[]>(
            `/sonar/zones/${resolvedZoneId}/scenarios`
          ),
          apiClient.get<ExpositionRecord[]>(
            `/sonar/zones/${resolvedZoneId}/expositions`
          ),
          apiClient.get<MonsterEncounterRecord[]>(
            `/sonar/zones/${resolvedZoneId}/monster-encounters`
          ),
          apiClient.get<ChallengeRecord[]>(
            `/sonar/zones/${resolvedZoneId}/challenges`
          ),
        ]);
        if (!active) return;
        setCharacters(fetchedCharacters);
        setTreasureChests(fetchedTreasureChests);
        setHealingFountains(fetchedHealingFountains);
        setScenarios(fetchedScenarios);
        setExpositions(fetchedExpositions);
        setMonsterEncounters(fetchedMonsterEncounters);
        setChallenges(fetchedChallenges);
        setZoneMapPinsError(null);
      } catch (error) {
        console.error('Error fetching zone map pins:', error);
        if (!active) return;
        setCharacters([]);
        setTreasureChests([]);
        setHealingFountains([]);
        setScenarios([]);
        setExpositions([]);
        setMonsterEncounters([]);
        setChallenges([]);
        setZoneMapPinsError(
          'Unable to load the full in-game pin set for this zone.'
        );
      }
    };

    const fetchZoneFlavorJobs = async () => {
      try {
        const response = await apiClient.get<ZoneFlavorGenerationJob[]>(
          `/sonar/admin/zone-flavor-generation-jobs?zoneId=${resolvedZoneId}&limit=10`
        );
        if (!active) return;
        setZoneFlavorJobs(response);
        setZoneFlavorJobsError(null);
      } catch (error) {
        console.error('Error fetching zone flavor jobs:', error);
        if (!active) return;
        setZoneFlavorJobsError('Unable to load zone flavor jobs.');
      }
    };

    const fetchZoneTagJobs = async () => {
      try {
        const response = await apiClient.get<ZoneTagGenerationJob[]>(
          `/sonar/admin/zone-tag-generation-jobs?zoneId=${resolvedZoneId}&limit=10`
        );
        if (!active) return;
        setZoneTagJobs(response);
        setZoneTagJobsError(null);
      } catch (error) {
        console.error('Error fetching zone tag jobs:', error);
        if (!active) return;
        setZoneTagJobsError('Unable to load zone tag jobs.');
      }
    };

    loadZoneMapPins();
    fetchZoneFlavorJobs();
    fetchZoneTagJobs();
    const interval = window.setInterval(() => {
      fetchZoneFlavorJobs();
      fetchZoneTagJobs();
    }, 5000);
    return () => {
      active = false;
      window.clearInterval(interval);
    };
  }, [apiClient, resolvedZoneId]);

  useEffect(() => {
    const latestJob = zoneFlavorJobs[0];
    if (!latestJob || latestJob.status !== 'completed') {
      return;
    }
    if (lastCompletedZoneFlavorJobIdRef.current === latestJob.id) {
      return;
    }
    lastCompletedZoneFlavorJobIdRef.current = latestJob.id;
    refreshZones();
  }, [zoneFlavorJobs, refreshZones]);

  useEffect(() => {
    const latestJob = zoneTagJobs[0];
    if (!latestJob || latestJob.status !== 'completed') {
      return;
    }
    if (lastCompletedZoneTagJobIdRef.current === latestJob.id) {
      return;
    }
    lastCompletedZoneTagJobIdRef.current = latestJob.id;
    refreshZones();
  }, [zoneTagJobs, refreshZones]);

  const handleMapClick = (lngLat: mapboxgl.LngLat) => {
    const newPoint: [number, number] = [lngLat.lng, lngLat.lat];
    setBoundaryPoints([...boundaryPoints, newPoint]);
  };

  const handleSaveBoundary = async () => {
    if (zone) {
      try {
        await apiClient.post(`/sonar/zones/${zone.id}/boundary`, {
          boundary: boundaryPoints,
        });
        setIsEditingBoundary(false);
      } catch (error) {
        console.error('Error saving boundary:', error);
      }
    }
  };

  const toggleBoundaryEditing = () => {
    setPinMoveError(null);
    setPinMoveStatus(null);
    setPinDeleteError(null);
    setPinDeleteStatus(null);
    setIsEditingBoundary((prev) => {
      const next = !prev;
      if (next) {
        setIsEditingPins(false);
      }
      return next;
    });
  };

  const togglePinEditing = () => {
    setPinMoveError(null);
    setPinMoveStatus(null);
    setPinDeleteError(null);
    setPinDeleteStatus(null);
    setIsEditingPins((prev) => {
      const next = !prev;
      if (next) {
        setIsEditingBoundary(false);
      }
      return next;
    });
  };

  const handlePinLocationChange = async (
    pin: AdminMapPin,
    coordinates: [number, number]
  ) => {
    const [longitude, latitude] = coordinates;
    setMovingPinId(pin.id);
    setPinMoveError(null);
    setPinMoveStatus(`Saving ${pin.name}...`);

    try {
      switch (pin.kind) {
        case 'pointOfInterest':
          await apiClient.patch(
            `/sonar/pointsOfInterest/${pin.entityId}/location`,
            { latitude, longitude }
          );
          break;
        case 'treasureChest':
          await apiClient.patch(
            `/sonar/treasure-chests/${pin.entityId}/location`,
            { latitude, longitude }
          );
          break;
        case 'healingFountain':
          await apiClient.patch(
            `/sonar/healing-fountains/${pin.entityId}/location`,
            { latitude, longitude }
          );
          break;
        case 'scenario':
          await apiClient.patch(`/sonar/scenarios/${pin.entityId}/location`, {
            latitude,
            longitude,
          });
          break;
        case 'exposition':
          await apiClient.patch(`/sonar/expositions/${pin.entityId}/location`, {
            latitude,
            longitude,
          });
          break;
        case 'monster':
          await apiClient.patch(
            `/sonar/monster-encounters/${pin.entityId}/location`,
            { latitude, longitude }
          );
          break;
        case 'challenge':
          await apiClient.patch(`/sonar/challenges/${pin.entityId}/location`, {
            latitude,
            longitude,
          });
          break;
        case 'character': {
          if (typeof pin.locationIndex !== 'number') {
            throw new Error(
              `Character pin ${pin.entityId} does not have a movable location index.`
            );
          }
          const character = characters.find(
            (entry) => entry.id === pin.entityId
          );
          if (!character) {
            throw new Error(`Character ${pin.entityId} was not found.`);
          }
          const currentLocations: CharacterLocation[] =
            character.locations ?? [];
          if (
            pin.locationIndex < 0 ||
            pin.locationIndex >= currentLocations.length
          ) {
            throw new Error(
              `Character ${pin.entityId} location index is out of bounds.`
            );
          }

          const nextLocations = currentLocations.map((location, index) =>
            index === pin.locationIndex
              ? { ...location, latitude, longitude }
              : location
          );

          await apiClient.put(`/sonar/characters/${pin.entityId}/locations`, {
            locations: nextLocations.map((location) => ({
              latitude: location.latitude,
              longitude: location.longitude,
            })),
          });

          setCharacters((prev) =>
            prev.map((entry) =>
              entry.id === pin.entityId
                ? { ...entry, locations: nextLocations }
                : entry
            )
          );
          break;
        }
        default:
          break;
      }

      setPinLocationOverrides((prev) => ({
        ...prev,
        [pin.id]: [longitude, latitude],
      }));
      setPinMoveStatus(`${pin.name} moved.`);
    } catch (error) {
      console.error('Error updating pin location:', error);
      setPinMoveError(`Unable to move ${pin.name} right now.`);
      throw error;
    } finally {
      setMovingPinId(null);
    }
  };

  const handleDeletePin = async (pin: AdminMapPin) => {
    const noun = pinDeleteLabelByKind[pin.kind];
    const confirmed = window.confirm(
      `Delete ${pin.name} and remove this ${noun} from the zone?`
    );
    if (!confirmed) return;

    setDeletingPinId(pin.id);
    setPinDeleteError(null);
    setPinDeleteStatus(`Deleting ${pin.name}...`);

    try {
      switch (pin.kind) {
        case 'pointOfInterest':
          await apiClient.delete(`/sonar/pointsOfInterest/${pin.entityId}`);
          setDeletedPointOfInterestIds((prev) => {
            const next = new Set(prev);
            next.add(pin.entityId);
            return next;
          });
          break;
        case 'treasureChest':
          await apiClient.delete(`/sonar/treasure-chests/${pin.entityId}`);
          setTreasureChests((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'healingFountain':
          await apiClient.delete(`/sonar/healing-fountains/${pin.entityId}`);
          setHealingFountains((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'scenario':
          await apiClient.delete(`/sonar/scenarios/${pin.entityId}`);
          setScenarios((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'exposition':
          await apiClient.delete(`/sonar/expositions/${pin.entityId}`);
          setExpositions((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'monster':
          await apiClient.delete(`/sonar/monster-encounters/${pin.entityId}`);
          setMonsterEncounters((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'challenge':
          await apiClient.delete(`/sonar/challenges/${pin.entityId}`);
          setChallenges((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        case 'character':
          await apiClient.delete(`/sonar/characters/${pin.entityId}`);
          setCharacters((prev) =>
            prev.filter((entry) => entry.id !== pin.entityId)
          );
          break;
        default:
          break;
      }

      setPinLocationOverrides((prev) => {
        if (!(pin.id in prev)) return prev;
        const next = { ...prev };
        delete next[pin.id];
        return next;
      });
      setPinDeleteStatus(`${pin.name} deleted.`);
    } catch (error) {
      console.error('Error deleting pin entity:', error);
      setPinDeleteError(`Unable to delete ${pin.name} right now.`);
    } finally {
      setDeletingPinId(null);
    }
  };

  const handleQueryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      setQuery(e.target.value);
      setShowPlaces(true);
    }, 500);
  };

  const handleCandidateSelect = (candidate: Candidate) => {
    if (!resolvedZoneId) {
      return;
    }
    importPointOfInterest(candidate.place_id, resolvedZoneId);
  };

  const handleGenerateQuest = async () => {
    if (!resolvedZoneId) {
      return;
    }
    await apiClient.post(`/sonar/zones/${resolvedZoneId}/questArchetypes`, {});
  };

  const handleGenerateZoneFlavor = async () => {
    if (!resolvedZoneId || queueingZoneFlavor) return;
    setQueueingZoneFlavor(true);
    try {
      const queuedJob = await apiClient.post<ZoneFlavorGenerationJob>(
        '/sonar/admin/zone-flavor-generation-jobs',
        { zoneId: resolvedZoneId }
      );
      setZoneFlavorJobs((prev) => [
        queuedJob,
        ...prev.filter((job) => job.id !== queuedJob.id),
      ]);
      setZoneFlavorJobsError(null);
    } catch (error) {
      console.error('Error queueing zone flavor job:', error);
      setZoneFlavorJobsError('Unable to queue zone flavor generation.');
    } finally {
      setQueueingZoneFlavor(false);
    }
  };

  const handleGenerateZoneTags = async () => {
    if (!resolvedZoneId || queueingZoneTags) return;
    setQueueingZoneTags(true);
    try {
      const queuedJob = await apiClient.post<ZoneTagGenerationJob>(
        '/sonar/admin/zone-tag-generation-jobs',
        { zoneId: resolvedZoneId }
      );
      setZoneTagJobs((prev) => [
        queuedJob,
        ...prev.filter((job) => job.id !== queuedJob.id),
      ]);
      setZoneTagJobsError(null);
    } catch (error) {
      console.error('Error queueing zone tag job:', error);
      setZoneTagJobsError('Unable to queue zone tag generation.');
    } finally {
      setQueueingZoneTags(false);
    }
  };

  const handleSaveZoneMetadata = async () => {
    if (!resolvedZoneId) {
      return;
    }
    setIsSavingZone(true);
    try {
      await apiClient.patch(`/sonar/zones/${resolvedZoneId}/edit`, {
        name,
        description,
        internalTags: parseInternalTagsInput(internalTagsInput),
      });
      await refreshZones();
      setIsEditingZone(false);
    } catch (error) {
      console.error('Error updating zone:', error);
      window.alert('Unable to save zone changes right now.');
    } finally {
      setIsSavingZone(false);
    }
  };

  const filteredPoints = pointsOfInterest.filter(
    (point) =>
      !deletedPointOfInterestIds.has(point.id) &&
      point.name.toLowerCase().includes(nameFilter.toLowerCase())
  );
  const latestZoneFlavorJob = zoneFlavorJobs[0];
  const latestZoneTagJob = zoneTagJobs[0];
  const allZoneBoundaries = zones
    .filter(
      (candidateZone) =>
        candidateZone.id !== resolvedZoneId && candidateZone.points.length > 0
    )
    .map((candidateZone) =>
      candidateZone.points.map(
        (point) => [point.longitude, point.latitude] as [number, number]
      )
    );
  const mapPins = useMemo<AdminMapPin[]>(() => {
    const resolveCoordinates = (
      pinId: string,
      longitude: number,
      latitude: number
    ): [number, number] => pinLocationOverrides[pinId] ?? [longitude, latitude];
    const sortedBoundaryPoints = sortBoundaryPoints(boundaryPoints);
    const zonePolygon =
      sortedBoundaryPoints.length >= 4
        ? turf.polygon([sortedBoundaryPoints])
        : null;
    const pointOfInterestById = new Map(
      pointsOfInterest
        .filter((point) => !deletedPointOfInterestIds.has(point.id))
        .map((point) => [point.id, point] as const)
    );

    const poiPins = pointsOfInterest
      .map((point) => {
        if (deletedPointOfInterestIds.has(point.id)) {
          return null;
        }
        const latitude = Number(point.lat);
        const longitude = Number(point.lng);
        if (!isValidCoordinate(latitude, longitude)) {
          return null;
        }

        return {
          id: `poi:${point.id}`,
          entityId: point.id,
          kind: 'pointOfInterest' as const,
          name: point.name || 'Point of Interest',
          coordinates: resolveCoordinates(
            `poi:${point.id}`,
            longitude,
            latitude
          ),
          imageUrl:
            point.thumbnailUrl?.trim() ||
            point.imageURL?.trim() ||
            markerStyleByKind.pointOfInterest.fallbackImage,
          draggable: true,
        };
      })
      .filter((pin): pin is AdminMapPin => pin !== null);

    const treasureChestPins = treasureChests
      .filter((chest) => isValidCoordinate(chest.latitude, chest.longitude))
      .map((chest) => ({
        id: `treasureChest:${chest.id}`,
        entityId: chest.id,
        kind: 'treasureChest' as const,
        name: chest.unlockTier
          ? `Treasure Chest T${chest.unlockTier}`
          : 'Treasure Chest',
        coordinates: resolveCoordinates(
          `treasureChest:${chest.id}`,
          chest.longitude,
          chest.latitude
        ),
        imageUrl: markerStyleByKind.treasureChest.fallbackImage,
        draggable: true,
      }));

    const healingFountainPins = healingFountains
      .filter((fountain) =>
        isValidCoordinate(fountain.latitude, fountain.longitude)
      )
      .map((fountain) => ({
        id: `healingFountain:${fountain.id}`,
        entityId: fountain.id,
        kind: 'healingFountain' as const,
        name: fountain.name || 'Healing Fountain',
        coordinates: resolveCoordinates(
          `healingFountain:${fountain.id}`,
          fountain.longitude,
          fountain.latitude
        ),
        imageUrl:
          fountain.thumbnailUrl?.trim() ||
          markerStyleByKind.healingFountain.fallbackImage,
        draggable: true,
      }));

    const scenarioPins = scenarios
      .filter((scenario) =>
        isValidCoordinate(scenario.latitude, scenario.longitude)
      )
      .map((scenario) => ({
        id: `scenario:${scenario.id}`,
        entityId: scenario.id,
        kind: 'scenario' as const,
        name: scenario.prompt || 'Scenario',
        coordinates: resolveCoordinates(
          `scenario:${scenario.id}`,
          scenario.longitude,
          scenario.latitude
        ),
        imageUrl:
          scenario.thumbnailUrl?.trim() ||
          scenario.imageUrl?.trim() ||
          markerStyleByKind.scenario.fallbackImage,
        draggable: true,
      }));

    const expositionPins = expositions
      .filter((exposition) =>
        isValidCoordinate(exposition.latitude, exposition.longitude)
      )
      .map((exposition) => ({
        id: `exposition:${exposition.id}`,
        entityId: exposition.id,
        kind: 'exposition' as const,
        name: exposition.title || 'Exposition',
        coordinates: resolveCoordinates(
          `exposition:${exposition.id}`,
          exposition.longitude,
          exposition.latitude
        ),
        imageUrl:
          exposition.thumbnailUrl?.trim() ||
          exposition.imageUrl?.trim() ||
          markerStyleByKind.exposition.fallbackImage,
        draggable: true,
      }));

    const monsterPins = monsterEncounters
      .filter((monster) =>
        isValidCoordinate(monster.latitude, monster.longitude)
      )
      .map((monster) => ({
        id: `monster:${monster.id}`,
        entityId: monster.id,
        kind: 'monster' as const,
        name: monster.name || 'Monster Encounter',
        coordinates: resolveCoordinates(
          `monster:${monster.id}`,
          monster.longitude,
          monster.latitude
        ),
        imageUrl:
          monster.thumbnailUrl?.trim() ||
          monster.imageUrl?.trim() ||
          markerStyleByKind.monster.fallbackImage,
        draggable: true,
      }));

    const challengePins = challenges
      .filter((challenge) =>
        isValidCoordinate(challenge.latitude, challenge.longitude)
      )
      .map((challenge) => ({
        id: `challenge:${challenge.id}`,
        entityId: challenge.id,
        kind: 'challenge' as const,
        name: challenge.question || 'Challenge',
        coordinates: resolveCoordinates(
          `challenge:${challenge.id}`,
          challenge.longitude,
          challenge.latitude
        ),
        imageUrl:
          challenge.thumbnailUrl?.trim() ||
          challenge.imageUrl?.trim() ||
          markerStyleByKind.challenge.fallbackImage,
        draggable: !(
          challenge.polygonPoints?.length && challenge.polygonPoints.length >= 3
        ),
      }));

    const characterPins = characters.flatMap((character) => {
      const pins: AdminMapPin[] = [];
      const imageUrl =
        character.thumbnailUrl?.trim() ||
        character.mapIconUrl?.trim() ||
        markerStyleByKind.character.fallbackImage;
      const characterName = character.name?.trim() || 'Character';

      if (character.pointOfInterestId) {
        const point = pointOfInterestById.get(character.pointOfInterestId);
        if (point) {
          const latitude = Number(point.lat);
          const longitude = Number(point.lng);
          if (isValidCoordinate(latitude, longitude)) {
            pins.push({
              id: `character:${character.id}:poi`,
              entityId: character.id,
              kind: 'character',
              name: characterName,
              coordinates: resolveCoordinates(
                `character:${character.id}:poi`,
                longitude,
                latitude
              ),
              imageUrl,
              draggable: false,
              dragHint: 'Location tied to point of interest',
            });
          }
        }
      }

      if (!zonePolygon) {
        return pins;
      }

      (character.locations ?? []).forEach((location, index) => {
        if (!isValidCoordinate(location.latitude, location.longitude)) {
          return;
        }
        const isInZone = turf.booleanPointInPolygon(
          turf.point([location.longitude, location.latitude]),
          zonePolygon
        );
        if (!isInZone) {
          return;
        }

        pins.push({
          id: `character:${character.id}:location:${index}`,
          entityId: character.id,
          kind: 'character',
          name:
            (character.locations?.length ?? 0) > 1
              ? `${characterName} #${index + 1}`
              : characterName,
          coordinates: resolveCoordinates(
            `character:${character.id}:location:${index}`,
            location.longitude,
            location.latitude
          ),
          imageUrl,
          draggable: true,
          locationIndex: index,
        });
      });

      return pins;
    });

    return [
      ...poiPins,
      ...characterPins,
      ...treasureChestPins,
      ...healingFountainPins,
      ...scenarioPins,
      ...expositionPins,
      ...monsterPins,
      ...challengePins,
    ];
  }, [
    pointsOfInterest,
    characters,
    treasureChests,
    healingFountains,
    scenarios,
    expositions,
    monsterEncounters,
    challenges,
    boundaryPoints,
    pinLocationOverrides,
    deletedPointOfInterestIds,
  ]);

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!zone) {
    return <div>Zone not found</div>;
  }

  if (generatePointsOfInterestLoading) {
    return <div>Generating points of interest...</div>;
  }
  if (generatePointsOfInterestError) {
    return <div>Error: {generatePointsOfInterestError.message}</div>;
  }

  return (
    <div className="m-10 p-8 bg-white rounded-lg shadow-lg">
      <h1 className="text-3xl font-bold mb-6 text-gray-800">{zone?.name}</h1>
      <p className="text-lg text-gray-600 mb-3">Latitude: {zone?.latitude}</p>
      <p className="text-lg text-gray-600 mb-3">Longitude: {zone?.longitude}</p>
      <div className="mb-6 rounded-lg border border-slate-200 bg-slate-50 px-4 py-3">
        <div className="text-sm font-semibold uppercase tracking-wide text-slate-500">
          Description
        </div>
        <p className="mt-2 whitespace-pre-wrap text-base text-slate-700">
          {zone.description?.trim() || 'No description yet.'}
        </p>
      </div>
      <div className="mb-6 rounded-lg border border-slate-200 bg-slate-50 px-4 py-3">
        <div className="text-sm font-semibold uppercase tracking-wide text-slate-500">
          Internal Tags
        </div>
        <p className="mt-2 whitespace-pre-wrap text-base text-slate-700">
          {(zone?.internalTags ?? []).join(', ') || 'None'}
        </p>
      </div>

      <div className="mb-6 space-x-2">
        <button
          onClick={() => setIsGenerating(!isGenerating)}
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
        >
          {isGenerating ? 'Stop Generating' : 'Generate Points of Interest'}
        </button>
        <button
          onClick={() => setIsImporting(!isImporting)}
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
        >
          Import Point of Interest
        </button>
        <button
          onClick={() => setIsGeneratingQuest(!isGeneratingQuest)}
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
        >
          Generate Quest
        </button>
        <button
          onClick={handleGenerateQuest}
          className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 disabled:bg-blue-300"
        >
          Generate Quests for Zone
        </button>
        <button
          onClick={() => setIsEditingZone(!isEditingZone)}
          className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 disabled:bg-blue-300"
        >
          Edit Zone
        </button>
        <button
          onClick={handleGenerateZoneFlavor}
          disabled={queueingZoneFlavor}
          className="bg-emerald-600 text-white px-4 py-2 rounded-md hover:bg-emerald-700 disabled:bg-emerald-300"
        >
          {queueingZoneFlavor ? 'Queueing Flavor...' : 'Generate Zone Flavor'}
        </button>
        <button
          onClick={handleGenerateZoneTags}
          disabled={queueingZoneTags}
          className="bg-violet-600 text-white px-4 py-2 rounded-md hover:bg-violet-700 disabled:bg-violet-300"
        >
          {queueingZoneTags ? 'Queueing Tags...' : 'Generate Zone Tags'}
        </button>
      </div>

      {(latestZoneFlavorJob || zoneFlavorJobsError) && (
        <div className="mb-6 rounded-lg border border-emerald-100 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          {latestZoneFlavorJob && (
            <>
              <div className="font-semibold">
                Zone flavor job: {latestZoneFlavorJob.status.replace(/_/g, ' ')}
              </div>
              {latestZoneFlavorJob.generatedDescription && (
                <p className="mt-2 text-emerald-800">
                  {latestZoneFlavorJob.generatedDescription}
                </p>
              )}
              {latestZoneFlavorJob.errorMessage && (
                <p className="mt-2 text-red-700">
                  {latestZoneFlavorJob.errorMessage}
                </p>
              )}
              {isZoneFlavorPendingStatus(latestZoneFlavorJob.status) && (
                <p className="mt-2 text-emerald-700">
                  The worker is generating a matched zone name and in-world
                  flavor from this zone’s boundary coordinates.
                </p>
              )}
            </>
          )}
          {zoneFlavorJobsError && (
            <p
              className={
                latestZoneFlavorJob ? 'mt-2 text-red-700' : 'text-red-700'
              }
            >
              {zoneFlavorJobsError}
            </p>
          )}
        </div>
      )}

      {(latestZoneTagJob || zoneTagJobsError) && (
        <div className="mb-6 rounded-lg border border-violet-100 bg-violet-50 px-4 py-3 text-sm text-violet-950">
          {latestZoneTagJob && (
            <>
              <div className="font-semibold">
                Zone tag job: {latestZoneTagJob.status.replace(/_/g, ' ')}
              </div>
              {latestZoneTagJob.generatedSummary && (
                <p className="mt-2 text-violet-900">
                  {latestZoneTagJob.generatedSummary}
                </p>
              )}
              {(latestZoneTagJob.selectedTags?.length ?? 0) > 0 && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {(latestZoneTagJob.selectedTags ?? []).map((tag) => (
                    <span
                      key={tag}
                      className="inline-flex items-center rounded-full border border-violet-200 bg-white px-3 py-1 text-xs font-medium text-violet-700"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
              {latestZoneTagJob.errorMessage && (
                <p className="mt-2 text-red-700">
                  {latestZoneTagJob.errorMessage}
                </p>
              )}
              {isZoneTagPendingStatus(latestZoneTagJob.status) && (
                <p className="mt-2 text-violet-800">
                  The worker is collecting neighborhood context from this
                  zone&apos;s geometry and points of interest, then selecting
                  five shared flavor tags.
                </p>
              )}
              {latestZoneTagJob.contextSnapshot && (
                <details className="mt-3">
                  <summary className="cursor-pointer text-xs font-semibold uppercase tracking-wide text-violet-700">
                    Context Used
                  </summary>
                  <pre className="mt-2 whitespace-pre-wrap rounded-md border border-violet-100 bg-white p-3 text-xs text-violet-900">
                    {latestZoneTagJob.contextSnapshot}
                  </pre>
                </details>
              )}
            </>
          )}
          {zoneTagJobsError && (
            <p
              className={
                latestZoneTagJob ? 'mt-2 text-red-700' : 'text-red-700'
              }
            >
              {zoneTagJobsError}
            </p>
          )}
        </div>
      )}

      {isEditingZone && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h2 className="text-xl font-semibold mb-4">Edit Zone</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Name
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Internal Tags
                </label>
                <input
                  type="text"
                  value={internalTagsInput}
                  onChange={(e) => setInternalTagsInput(e.target.value)}
                  placeholder="starter_region, downtown, high_density"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Comma-separated internal metadata tags. Not shown to players.
                </p>
              </div>
            </div>

            <div className="mt-6 flex justify-end space-x-3">
              <button
                onClick={() => setIsEditingZone(false)}
                className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveZoneMetadata}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                disabled={!resolvedZoneId || isSavingZone}
              >
                {isSavingZone ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Map Section */}
      <div className="mb-8">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Zone Boundary</h2>
          <div className="space-x-2">
            <button
              onClick={() => setBoundaryPoints([])}
              className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md"
            >
              Clear Boundary
            </button>
            <button
              onClick={toggleBoundaryEditing}
              className={`px-4 py-2 rounded-md ${
                isEditingBoundary
                  ? 'bg-red-500 hover:bg-red-600'
                  : 'bg-blue-500 hover:bg-blue-600'
              } text-white`}
            >
              {isEditingBoundary ? 'Stop Editing' : 'Edit Boundary'}
            </button>
            <button
              onClick={togglePinEditing}
              className={`px-4 py-2 rounded-md ${
                isEditingPins
                  ? 'bg-amber-600 hover:bg-amber-700'
                  : 'bg-slate-700 hover:bg-slate-800'
              } text-white`}
            >
              {isEditingPins ? 'Stop Moving Pins' : 'Move Pins'}
            </button>
            {isEditingBoundary && (
              <button
                onClick={handleSaveBoundary}
                className="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-md"
              >
                Save Boundary
              </button>
            )}
          </div>
        </div>
        <ZoneMap
          center={[zone.longitude, zone.latitude]}
          onMapClick={isEditingBoundary ? handleMapClick : undefined}
          boundaryPoints={boundaryPoints}
          allZoneBoundaries={allZoneBoundaries}
          showBoundaryControls={true}
          pins={mapPins}
          pinDragEnabled={isEditingPins}
          onPinLocationChange={handlePinLocationChange}
        />
        <div className="mt-3 flex flex-wrap gap-2 text-xs text-slate-600">
          {(
            [
              ['pointOfInterest', 'POI'],
              ['character', 'Character'],
              ['treasureChest', 'Treasure Chest'],
              ['healingFountain', 'Healing Fountain'],
              ['scenario', 'Scenario'],
              ['exposition', 'Exposition'],
              ['monster', 'Monster'],
              ['challenge', 'Challenge'],
            ] as Array<[AdminMapPinKind, string]>
          ).map(([kind, label]) => (
            <span
              key={kind}
              className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1"
            >
              <span
                className="inline-block h-2.5 w-2.5 rounded-full"
                style={{ backgroundColor: markerStyleByKind[kind].ring }}
              />
              {label}
            </span>
          ))}
        </div>
        {zoneMapPinsError ? (
          <p className="mt-2 text-sm text-amber-700">{zoneMapPinsError}</p>
        ) : null}
        {pinMoveError ? (
          <p className="mt-2 text-sm text-red-600">{pinMoveError}</p>
        ) : null}
        {pinMoveStatus ? (
          <p className="mt-2 text-sm text-slate-600">
            {movingPinId ? 'Saving location...' : pinMoveStatus}
          </p>
        ) : null}
        {isEditingPins && !pinMoveError ? (
          <p className="text-sm text-gray-600 mt-2">
            Drag pins to reposition them. Standalone pins save on drop. Polygon
            challenges stay locked to their polygon editor.
          </p>
        ) : null}
        {pinDeleteError ? (
          <p className="mt-2 text-sm text-red-600">{pinDeleteError}</p>
        ) : null}
        {pinDeleteStatus ? (
          <p className="mt-2 text-sm text-slate-600">
            {deletingPinId ? 'Deleting pin...' : pinDeleteStatus}
          </p>
        ) : null}
        {isEditingBoundary && (
          <p className="text-sm text-gray-600 mt-2">
            Click on the map to add boundary points. The gameplay pins stay
            visible for context, but they won&apos;t intercept clicks while
            boundary editing is on.
          </p>
        )}
        <div className="mt-5 rounded-lg border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h3 className="text-sm font-semibold uppercase tracking-wide text-slate-600">
                Zone Pins
              </h3>
              <p className="mt-1 text-sm text-slate-500">
                Delete a pin here to remove the underlying entity from this
                zone.
              </p>
            </div>
            <div className="text-sm text-slate-500">{mapPins.length} total</div>
          </div>
          {mapPins.length === 0 ? (
            <p className="mt-4 text-sm text-slate-500">
              No in-zone pins found.
            </p>
          ) : (
            <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {mapPins.map((pin) => (
                <div
                  key={pin.id}
                  className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm"
                >
                  <div className="flex items-start gap-3">
                    <img
                      src={
                        pin.imageUrl ||
                        markerStyleByKind[pin.kind].fallbackImage
                      }
                      alt={pin.name}
                      className="h-12 w-12 rounded-full border object-cover"
                    />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span
                          className="inline-block h-2.5 w-2.5 rounded-full"
                          style={{
                            backgroundColor: markerStyleByKind[pin.kind].ring,
                          }}
                        />
                        <div className="truncate font-medium text-slate-900">
                          {pin.name}
                        </div>
                      </div>
                      <div className="mt-1 text-xs uppercase tracking-wide text-slate-500">
                        {markerStyleByKind[pin.kind].fullLabel}
                      </div>
                      <div className="mt-2 text-xs text-slate-500">
                        {pin.coordinates[1].toFixed(5)},{' '}
                        {pin.coordinates[0].toFixed(5)}
                      </div>
                      {!pin.draggable ? (
                        <div className="mt-1 text-xs text-amber-700">
                          {pin.dragHint ??
                            'Location locked to polygon geometry'}
                        </div>
                      ) : null}
                    </div>
                  </div>
                  <div className="mt-3 flex justify-end">
                    <button
                      type="button"
                      onClick={() => void handleDeletePin(pin)}
                      disabled={deletingPinId === pin.id}
                      className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      {deletingPinId === pin.id ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="mb-4">
        <input
          type="text"
          placeholder="Filter points by name..."
          value={nameFilter}
          onChange={(e) => setNameFilter(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredPoints.map((point) => (
          <div key={point.id} className="bg-gray-100 p-6 rounded-lg shadow-md">
            {point.imageURL && (
              <>
                <div className="relative">
                  <img
                    src={point.imageURL}
                    alt={point.name}
                    className="float-left mr-4 mb-4 w-32 h-32 object-cover rounded cursor-pointer"
                    onClick={() => setSelectedImage(point.imageURL)}
                  />
                </div>
                {selectedImage === point.imageURL && (
                  <div
                    className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
                    onClick={() => setSelectedImage(null)}
                  >
                    <div className="relative">
                      <img
                        src={selectedImage}
                        alt={point.name}
                        className="max-h-[90vh] max-w-[90vw] object-contain"
                      />
                      <button
                        className="absolute top-4 right-4 text-white text-xl font-bold"
                        onClick={() => setSelectedImage(null)}
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}
            <h2 className="text-xl font-bold mb-2 text-gray-800">
              {point.name}
            </h2>
            <p className="text-gray-600 mb-3">
              Description: {point.description}
            </p>
            <p className="text-gray-600 mb-3">Type: {point.originalName}</p>
            <p className="text-gray-600 mb-3">
              Tags: {point.tags?.map((tag) => tag.name).join(', ') || 'No tags'}
            </p>
            <p className="text-gray-600 mb-3">Latitude: {point.lat}</p>
            <p className="text-gray-600 mb-3">Longitude: {point.lng}</p>
            <p className="text-gray-600 mb-3">Clue: {point.clue}</p>
            <p className="text-gray-600 mb-3">
              Created At: {point.createdAt.toLocaleString()}
            </p>
            <p className="text-gray-600 mb-3">
              Updated At: {point.updatedAt.toLocaleString()}
            </p>
            <p className="text-gray-600 mb-3">
              Place ID:{' '}
              <a
                href={`/place/${point.googleMapsPlaceId}`}
                className="text-blue-500 hover:underline"
              >
                {point.googleMapsPlaceId}
              </a>
            </p>
            <button
              className="bg-blue-500 hover:bg-blue-600 text-white rounded-md px-3 py-1 text-sm font-medium shadow-md mr-2 transition duration-200"
              onClick={(e) => {
                e.stopPropagation();
                refreshPointOfInterestImage(point.id);
              }}
            >
              Refresh Image
            </button>
            <button
              className="bg-green-500 hover:bg-green-600 text-white rounded-md px-3 py-1 text-sm font-medium shadow-md transition duration-200"
              onClick={(e) => {
                e.stopPropagation();
                refreshPointOfInterest(point.id);
              }}
            >
              Refresh POI
            </button>
          </div>
        ))}
      </div>

      {isImporting && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">Import Point of Interest</h2>
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="importText"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Enter Import Text
                </label>
                <input
                  id="importText"
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  onChange={handleQueryChange}
                />
                {showPlaces && (
                  <div className="w-full bg-white border rounded-lg shadow-lg max-h-48 overflow-y-auto mt-1">
                    {candidatesLoading && (
                      <div className="px-4 py-2 text-sm text-gray-500">
                        Searching...
                      </div>
                    )}
                    {candidatesError && (
                      <div className="px-4 py-2 text-sm text-red-600">
                        {candidatesError.message}
                      </div>
                    )}
                    {!candidatesLoading && candidates.length === 0 && (
                      <div className="px-4 py-2 text-sm text-gray-500">
                        No results yet.
                      </div>
                    )}
                    {candidates.map((candidate, index) => (
                      <div
                        key={`${candidate.place_id}-${index}`}
                        className="px-4 py-2 hover:bg-gray-100 cursor-pointer"
                        onClick={() => handleCandidateSelect(candidate)}
                      >
                        <div className="font-medium">{candidate.name}</div>
                        <div className="text-xs text-gray-500">
                          {candidate.formatted_address}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {isGeneratingQuest && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">Generate Quest</h2>
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="questType"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Quest Archetype
                </label>
                <select
                  id="questType"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 [&>*]:w-full"
                  onChange={(e) => setSelectedQuestArchtype(e.target.value)}
                >
                  {questArchtypes.map((questArchtype) => (
                    <option key={questArchtype.id} value={questArchtype.id}>
                      {questArchtype.id}
                    </option>
                  ))}
                </select>
              </div>
              <button
                disabled={!selectedQuestArchtype}
                onClick={() => {
                  if (selectedQuestArchtype) {
                    if (!resolvedZoneId) {
                      return;
                    }
                    generateQuest(resolvedZoneId, selectedQuestArchtype);
                    setIsGeneratingQuest(false);
                  }
                }}
                className="w-full bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600"
              >
                Generate Quest
              </button>
              <button
                onClick={() => setIsGeneratingQuest(false)}
                className="w-full bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {isGenerating && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">
              Generate Points of Interest
            </h2>

            {placeTypesLoading ? (
              <p>Loading place types...</p>
            ) : placeTypesError ? (
              <p className="text-red-500">
                Error loading place types: {placeTypesError.message}
              </p>
            ) : (
              <div className="space-y-4">
                <div>
                  <div className="mb-4">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">
                      Selected Included Types:
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {selectedIncludedPlaceTypes.map((type) => (
                        <span
                          key={type}
                          className="px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-sm"
                        >
                          {type}
                        </span>
                      ))}
                      {selectedIncludedPlaceTypes.length === 0 && (
                        <span className="text-gray-500 text-sm">
                          No types selected
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="mb-4">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">
                      Selected Excluded Types:
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {selectedExcludedPlaceTypes.map((type) => (
                        <span
                          key={type}
                          className="px-2 py-1 bg-red-100 text-red-800 rounded-md text-sm"
                        >
                          {type}
                        </span>
                      ))}
                      {selectedExcludedPlaceTypes.length === 0 && (
                        <span className="text-gray-500 text-sm">
                          No types selected
                        </span>
                      )}
                    </div>
                  </div>
                  <label
                    htmlFor="placeType"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Included Place Types
                  </label>
                  <select
                    id="placeType"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    multiple
                    value={selectedIncludedPlaceTypes}
                    onChange={(e) => {
                      const selectedOptions = Array.from(
                        e.target.selectedOptions,
                        (option) => option.value
                      );
                      setSelectedIncludedPlaceTypes(selectedOptions);
                    }}
                  >
                    {placeTypes.map((type) => (
                      <option key={type} value={type}>
                        {type}
                      </option>
                    ))}
                  </select>
                  <label
                    htmlFor="excludedPlaceTypes"
                    className="block text-sm font-medium text-gray-700 mb-1 mt-4"
                  >
                    Excluded Place Types
                  </label>
                  <select
                    id="excludedPlaceTypes"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    multiple
                    value={selectedExcludedPlaceTypes}
                    onChange={(e) => {
                      const selectedOptions = Array.from(
                        e.target.selectedOptions,
                        (option) => option.value
                      );
                      setSelectedExcludedPlaceTypes(selectedOptions);
                    }}
                  >
                    {placeTypes.map((type) => (
                      <option key={type} value={type}>
                        {type}
                      </option>
                    ))}
                  </select>
                  <label
                    htmlFor="numPlaces"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Number of Places
                  </label>
                  <input
                    id="numPlaces"
                    type="number"
                    min="1"
                    max="20"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    value={numPlaces}
                    onChange={(e) =>
                      setNumPlaces(
                        Math.min(20, Math.max(1, parseInt(e.target.value)))
                      )
                    }
                  />
                </div>

                {generatePointsOfInterestError && (
                  <p className="text-red-500">
                    Error: {generatePointsOfInterestError}
                  </p>
                )}

                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => setIsGenerating(false)}
                    className="bg-gray-200 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-300"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      if (
                        selectedIncludedPlaceTypes &&
                        selectedExcludedPlaceTypes
                      ) {
                        generatePointsOfInterest(
                          resolvedZoneId,
                          selectedIncludedPlaceTypes,
                          selectedExcludedPlaceTypes,
                          numPlaces
                        );
                        setIsGenerating(false);
                      }
                    }}
                    disabled={
                      !selectedIncludedPlaceTypes || !selectedExcludedPlaceTypes
                    }
                    className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 disabled:bg-blue-300"
                  >
                    Generate
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
