import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useTagContext, useZoneContext } from '@poltergeist/contexts';
import { Candidate, Character, InventoryItem, LocationArchetype, PointOfInterest, Quest, QuestNode, QuestNodeChallenge, Tag } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import * as wellknown from 'wellknown';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useCandidates } from '@poltergeist/hooks';

type PointOfInterestImport = {
  id: string;
  placeId: string;
  zoneId: string;
  status: string;
  errorMessage?: string | null;
  pointOfInterestId?: string | null;
  createdAt: string;
  updatedAt: string;
};

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

const emptyQuestForm = {
  name: '',
  description: '',
  imageUrl: '',
  zoneId: '',
  questGiverCharacterId: '',
  questArchetypeId: '',
  gold: 0,
  itemRewards: [] as { inventoryItemId: string; quantity: number }[],
};

const emptyNodeForm = {
  orderIndex: 1,
  nodeType: 'poi' as 'poi' | 'polygon',
  pointOfInterestId: '',
  polygonPoints: '',
};

const emptyChallengeForm = {
  tier: 1,
  question: '',
  reward: 0,
  inventoryItemId: '',
  locationArchetypeId: '',
  locationChallenge: '',
};

const emptyQuestReward = {
  inventoryItemId: '',
  quantity: 1,
};

const parsePolygonPoints = (input: string): [number, number][] | null => {
  if (!input.trim()) return null;
  try {
    const parsed = JSON.parse(input);
    if (!Array.isArray(parsed)) return null;
    const points: [number, number][] = [];
    for (const entry of parsed) {
      if (!Array.isArray(entry) || entry.length < 2) return null;
      const lng = Number(entry[0]);
      const lat = Number(entry[1]);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
      points.push([lng, lat]);
    }
    return points.length ? points : null;
  } catch {
    return null;
  }
};

const parsePolygonWkt = (raw: string): number[][][] | null => {
  if (!raw) return null;
  let trimmed = raw.trim();
  if (!trimmed) return null;
  if (trimmed.toUpperCase().startsWith('SRID=')) {
    const parts = trimmed.split(';');
    trimmed = parts[parts.length - 1] || trimmed;
  }
  try {
    const geometry = wellknown.parse(trimmed);
    if (!geometry || geometry.type !== 'Polygon') return null;
    return geometry.coordinates as number[][][];
  } catch {
    return null;
  }
};

const normalizeText = (value: string) => value.trim().toLowerCase();

const closePolygonRing = (ring: [number, number][]) => {
  if (ring.length === 0) return ring;
  const [firstLng, firstLat] = ring[0];
  const [lastLng, lastLat] = ring[ring.length - 1];
  if (firstLng === lastLng && firstLat === lastLat) return ring;
  return [...ring, ring[0]];
};

const normalizePolygonCoordinates = (coords: number[][][] | null) => {
  if (!coords || coords.length === 0) return null;
  const ring = coords[0] ?? [];
  if (ring.length < 3) return null;
  return [closePolygonRing(ring as [number, number][])] as number[][][];
};

const matchLocationArchetypeForPoi = (
  poi: PointOfInterest,
  archetypes: LocationArchetype[]
): LocationArchetype | null => {
  if (!poi.tags || poi.tags.length === 0) return null;
  const tagSet = new Set(poi.tags.map((tag) => normalizeText(tag.name)));
  let best: LocationArchetype | null = null;
  let bestScore = 0;

  archetypes.forEach((archetype) => {
    const included = (archetype.includedTypes || []).map(normalizeText);
    let score = 0;
    included.forEach((type) => {
      if (tagSet.has(type)) score += 1;
    });
    if (score > bestScore) {
      bestScore = score;
      best = archetype;
    }
  });

  return bestScore > 0 ? best : null;
};

export const Quests = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { tagGroups } = useTagContext();
  const { locationArchetypes } = useQuestArchetypes();
  const [quests, setQuests] = useState<Quest[]>([]);
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [zoneSearch, setZoneSearch] = useState('');
  const [characterSearch, setCharacterSearch] = useState('');
  const [poiSearch, setPoiSearch] = useState('');
  const [poiFiltersOpen, setPoiFiltersOpen] = useState(false);
  const [poiZoneFilterId, setPoiZoneFilterId] = useState('');
  const [poiTagSearch, setPoiTagSearch] = useState('');
  const [poiTagFilterIds, setPoiTagFilterIds] = useState<string[]>([]);
  const [zonePoiMap, setZonePoiMap] = useState<Record<string, Set<string>>>({});
  const [zonePoiMapLoaded, setZonePoiMapLoaded] = useState(false);
  const [zonePoiMapLoading, setZonePoiMapLoading] = useState(false);
  const [zoneDetailsById, setZoneDetailsById] = useState<Record<string, { boundary?: number[][]; boundaryCoords?: { latitude: number; longitude: number }[]; latitude?: number; longitude?: number }>>({});
  const [selectedQuestId, setSelectedQuestId] = useState<string>('');
  const [showCreateQuest, setShowCreateQuest] = useState(false);
  const [questForm, setQuestForm] = useState({ ...emptyQuestForm });
  const [nodeForm, setNodeForm] = useState({ ...emptyNodeForm });
  const [polygonDraftPoints, setPolygonDraftPoints] = useState<[number, number][]>([]);
  const [challengeDrafts, setChallengeDrafts] = useState<Record<string, typeof emptyChallengeForm>>({});
  const [selectedPoiForModal, setSelectedPoiForModal] = useState<PointOfInterest | null>(null);
  const [characterLocationsOpen, setCharacterLocationsOpen] = useState(false);
  const [selectedCharacterLocations, setSelectedCharacterLocations] = useState<{ latitude: number; longitude: number }[]>([]);
  const [characterLocationsLoading, setCharacterLocationsLoading] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importQuery, setImportQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(null);
  const [importZoneId, setImportZoneId] = useState('');
  const [importError, setImportError] = useState<string | null>(null);
  const [importJobs, setImportJobs] = useState<PointOfInterestImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const { candidates } = useCandidates(importQuery);
  const [importToasts, setImportToasts] = useState<string[]>([]);
  const [notifiedImportIds, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const [polygonRefreshNonce, setPolygonRefreshNonce] = useState(0);
  const questMapContainer = useRef<HTMLDivElement>(null);
  const questMap = useRef<mapboxgl.Map | null>(null);
  const [questMapLoaded, setQuestMapLoaded] = useState(false);
  const questNodeMarkers = useRef<mapboxgl.Marker[]>([]);
  const poiMarkers = useRef<mapboxgl.Marker[]>([]);
  const characterLocationMarkers = useRef<mapboxgl.Marker[]>([]);

  const selectedQuest = useMemo(() => quests.find((quest) => quest.id === selectedQuestId) ?? null, [quests, selectedQuestId]);

  useEffect(() => {
    let isMounted = true;
    const load = async () => {
      const results = await Promise.allSettled([
        apiClient.get<Quest[]>('/sonar/quests'),
        apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest'),
        apiClient.get<Character[]>('/sonar/characters'),
        apiClient.get<InventoryItem[]>('/sonar/inventory-items'),
      ]);

      if (!isMounted) return;

      const [questsResult, poiResult, charactersResult, inventoryResult] = results;
      if (questsResult.status === 'fulfilled') {
        setQuests(questsResult.value);
      } else {
        console.error('Failed to load quests', questsResult.reason);
        setLoadError('Failed to load quests. Check console for details.');
      }

      if (poiResult.status === 'fulfilled') {
        setPointsOfInterest(poiResult.value);
      } else {
        console.error('Failed to load points of interest', poiResult.reason);
      }

      if (charactersResult.status === 'fulfilled') {
        setCharacters(charactersResult.value);
      } else {
        console.error('Failed to load characters', charactersResult.reason);
      }

      if (inventoryResult.status === 'fulfilled') {
        setInventoryItems(inventoryResult.value);
      } else {
        console.error('Failed to load inventory items', inventoryResult.reason);
      }

      setLoading(false);
    };

    load();
    return () => {
      isMounted = false;
    };
  }, []);

  const refreshPointsOfInterest = async () => {
    try {
      const response = await apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest');
      setPointsOfInterest(response);
    } catch (error) {
      console.error('Failed to refresh points of interest', error);
    }
  };

  useEffect(() => {
    if (!selectedQuest) return;
    const nextIndex = (selectedQuest.nodes?.length ?? 0) + 1;
    setNodeForm((prev) => ({ ...prev, orderIndex: nextIndex }));
  }, [selectedQuest]);

  useEffect(() => {
    if (!questForm.questGiverCharacterId) {
      setSelectedCharacterLocations([]);
      return;
    }
    let isMounted = true;
    const loadLocations = async () => {
      try {
        const response = await apiClient.get<{ latitude: number; longitude: number }[]>(
          `/sonar/characters/${questForm.questGiverCharacterId}/locations`
        );
        if (!isMounted) return;
        setSelectedCharacterLocations(response);
      } catch (error) {
        console.error('Failed to load character locations for map', error);
      }
    };
    loadLocations();
    return () => {
      isMounted = false;
    };
  }, [apiClient, questForm.questGiverCharacterId]);

  useEffect(() => {
    if (!questForm.zoneId) return;
    if (zoneDetailsById[questForm.zoneId]) return;
    let isMounted = true;
    const loadZoneDetail = async () => {
      try {
        const zone = await apiClient.get<{ boundary?: number[][]; boundaryCoords?: { latitude: number; longitude: number }[]; latitude?: number; longitude?: number }>(
          `/sonar/zones/${questForm.zoneId}`
        );
        if (!isMounted) return;
        console.log('Quest Map: loaded zone details', zone);
        setZoneDetailsById((prev) => ({ ...prev, [questForm.zoneId]: zone }));
      } catch (error) {
        console.error('Failed to load zone details', error);
      }
    };
    loadZoneDetail();
    return () => {
      isMounted = false;
    };
  }, [apiClient, questForm.zoneId, zoneDetailsById]);

  useEffect(() => {
    if (!poiFiltersOpen || zonePoiMapLoaded || zonePoiMapLoading || zones.length === 0) return;
    let isMounted = true;
    const loadZonePoiMap = async () => {
      setZonePoiMapLoading(true);
      const results = await Promise.allSettled(
        zones.map((zone) => apiClient.get<PointOfInterest[]>(`/sonar/zones/${zone.id}/pointsOfInterest`))
      );
      if (!isMounted) return;
      const nextMap: Record<string, Set<string>> = {};
      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          nextMap[zones[index].id] = new Set(result.value.map((poi) => poi.id));
        }
      });
      setZonePoiMap(nextMap);
      setZonePoiMapLoaded(true);
      setZonePoiMapLoading(false);
    };

    loadZonePoiMap();
    return () => {
      isMounted = false;
    };
  }, [apiClient, poiFiltersOpen, zonePoiMapLoaded, zonePoiMapLoading, zones]);

  useEffect(() => {
    if (questMapContainer.current && !questMap.current) {
      questMap.current = new mapboxgl.Map({
        container: questMapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [0, 0],
        zoom: 2,
        interactive: true,
      });

      questMap.current.on('load', () => {
        setQuestMapLoaded(true);
      });
    }

    return () => {
      if (questMap.current) {
        questMap.current.remove();
        questMap.current = null;
      }
    };
  }, [selectedQuest]);

  useEffect(() => {
    if (questMap.current && questMapLoaded) {
      questMap.current.resize();
    }
  }, [questMapLoaded, selectedQuest]);

  const questPolygons = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return [];
    return selectedQuest.nodes
      .filter((node) => node.polygon || (node.polygonPoints && node.polygonPoints.length >= 3))
      .map((node) => ({
        id: node.id,
        orderIndex: node.orderIndex,
        coordinates: normalizePolygonCoordinates(
          node.polygonPoints && node.polygonPoints.length >= 3
            ? [node.polygonPoints]
            : parsePolygonWkt(node.polygon ?? '')
        ),
      }))
      .filter((entry) => entry.coordinates);
  }, [selectedQuest?.nodes]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    if (!map.getSource('quest-node-draft-line')) {
      map.addSource('quest-node-draft-line', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'LineString', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'quest-node-draft-line',
        type: 'line',
        source: 'quest-node-draft-line',
        paint: {
          'line-color': '#0f766e',
          'line-width': 2,
          'line-dasharray': [2, 2],
        },
      });
    }

    if (!map.getSource('quest-node-draft-polygon')) {
      map.addSource('quest-node-draft-polygon', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'quest-node-draft-polygon',
        type: 'fill',
        source: 'quest-node-draft-polygon',
        paint: {
          'fill-color': '#14b8a6',
          'fill-opacity': 0.15,
        },
      });
      map.addLayer({
        id: 'quest-node-draft-polygon-outline',
        type: 'line',
        source: 'quest-node-draft-polygon',
        paint: {
          'line-color': '#0f766e',
          'line-width': 2,
        },
      });
    }

    if (!map.getSource('quest-node-polygons')) {
      map.addSource('quest-node-polygons', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: [],
        },
      });
      map.addLayer({
        id: 'quest-node-polygons-fill',
        type: 'fill',
        source: 'quest-node-polygons',
        paint: {
          'fill-color': '#f59e0b',
          'fill-opacity': 0.18,
        },
      });
      map.addLayer({
        id: 'quest-node-polygons-outline',
        type: 'line',
        source: 'quest-node-polygons',
        paint: {
          'line-color': '#b45309',
          'line-width': 2,
        },
      });
    }
  }, [questMapLoaded]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const lineSource = map.getSource('quest-node-draft-line') as mapboxgl.GeoJSONSource | undefined;
    const polygonSource = map.getSource('quest-node-draft-polygon') as mapboxgl.GeoJSONSource | undefined;

    const lineCoords = polygonDraftPoints;
    if (lineSource) {
      lineSource.setData({
        type: 'Feature',
        geometry: { type: 'LineString', coordinates: lineCoords },
        properties: {},
      });
    }

    if (polygonSource) {
      if (polygonDraftPoints.length >= 3) {
        const ring = [...polygonDraftPoints, polygonDraftPoints[0]];
        polygonSource.setData({
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [ring] },
          properties: {},
        });
      } else {
        polygonSource.setData({
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [] },
          properties: {},
        });
      }
    }
  }, [polygonDraftPoints, questMapLoaded]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const ensurePolygonSource = () => {
      if (!map.isStyleLoaded()) {
        return undefined;
      }
      let polygonSource = map.getSource('quest-node-polygons') as mapboxgl.GeoJSONSource | undefined;
      if (!polygonSource) {
        map.addSource('quest-node-polygons', {
          type: 'geojson',
          data: {
            type: 'FeatureCollection',
            features: [],
          },
        });
        map.addLayer({
          id: 'quest-node-polygons-fill',
          type: 'fill',
          source: 'quest-node-polygons',
          paint: {
            'fill-color': '#f59e0b',
            'fill-opacity': 0.18,
          },
        });
        map.addLayer({
          id: 'quest-node-polygons-outline',
          type: 'line',
          source: 'quest-node-polygons',
          paint: {
            'line-color': '#b45309',
            'line-width': 2,
          },
        });
        polygonSource = map.getSource('quest-node-polygons') as mapboxgl.GeoJSONSource | undefined;
      }
      return polygonSource;
    };

    if (!map.isStyleLoaded()) {
      const handleStyleLoad = () => {
        setPolygonRefreshNonce((prev) => prev + 1);
      };
      map.once('style.load', handleStyleLoad);
      return () => {
        map.off('style.load', handleStyleLoad);
      };
    }

    const polygonSource = ensurePolygonSource();
    if (!polygonSource) return;

    const features = questPolygons
      .filter((entry) => entry.coordinates && entry.coordinates.length > 0)
      .map((entry) => ({
        type: 'Feature' as const,
        properties: {
          id: entry.id,
          orderIndex: entry.orderIndex,
        },
        geometry: {
          type: 'Polygon' as const,
          coordinates: entry.coordinates ?? [],
        },
      }));

    console.log('Quest Map: polygon refresh', {
      totalNodes: selectedQuest?.nodes?.length ?? 0,
      polygonCount: questPolygons.length,
      features,
    });

    polygonSource.setData({
      type: 'FeatureCollection',
      features,
    });
  }, [questPolygons, questMapLoaded, polygonRefreshNonce, selectedQuest?.nodes?.length]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const handleClick = (event: mapboxgl.MapMouseEvent & mapboxgl.EventData) => {
      if (nodeForm.nodeType !== 'polygon') return;
      const { lng, lat } = event.lngLat;
      setPolygonDraftPoints((prev) => {
        const next = [...prev, [lng, lat]] as [number, number][];
        setNodeForm((formPrev) => ({
          ...formPrev,
          polygonPoints: JSON.stringify(next),
        }));
        return next;
      });
    };
    map.on('click', handleClick);
    return () => {
      map.off('click', handleClick);
    };
  }, [nodeForm.nodeType, questMapLoaded]);

  useEffect(() => {
    if (nodeForm.nodeType !== 'polygon') {
      setPolygonDraftPoints([]);
      setNodeForm((prev) => ({ ...prev, polygonPoints: '' }));
    }
  }, [nodeForm.nodeType]);


  const filteredQuests = useMemo(() => {
    if (!searchQuery.trim()) return quests;
    const term = searchQuery.toLowerCase();
    return quests.filter((quest) => quest.name.toLowerCase().includes(term));
  }, [quests, searchQuery]);

  const filteredZones = useMemo(() => {
    if (!zoneSearch.trim()) return zones;
    const term = zoneSearch.toLowerCase();
    return zones.filter((zone) => zone.name?.toLowerCase().includes(term));
  }, [zones, zoneSearch]);

  const filteredCharacters = useMemo(() => {
    if (!characterSearch.trim()) return characters;
    const term = characterSearch.toLowerCase();
    return characters.filter((character) => character.name?.toLowerCase().includes(term));
  }, [characters, characterSearch]);

  const allTags = useMemo(() => {
    const tags: Tag[] = [];
    const seen = new Set<string>();
    tagGroups.forEach((group) => {
      group.tags?.forEach((tag) => {
        if (!seen.has(tag.id)) {
          seen.add(tag.id);
          tags.push(tag);
        }
      });
    });
    return tags;
  }, [tagGroups]);

  const filteredTags = useMemo(() => {
    if (!poiTagSearch.trim()) return allTags;
    const term = poiTagSearch.toLowerCase();
    return allTags.filter((tag) => tag.name?.toLowerCase().includes(term));
  }, [allTags, poiTagSearch]);

  const filteredPointsOfInterest = useMemo(() => {
    let filtered = pointsOfInterest;
    if (poiSearch.trim()) {
      const term = poiSearch.toLowerCase();
      filtered = filtered.filter((poi) => {
        const name = poi.name?.toLowerCase() ?? '';
        const googleName = poi.googleMapsPlaceName?.toLowerCase() ?? '';
        const originalName = poi.originalName?.toLowerCase() ?? '';
        return name.includes(term) || googleName.includes(term) || originalName.includes(term);
      });
    }
    if (poiZoneFilterId && zonePoiMap[poiZoneFilterId]) {
      const allowed = zonePoiMap[poiZoneFilterId];
      filtered = filtered.filter((poi) => allowed.has(poi.id));
    }
    if (poiTagFilterIds.length > 0) {
      filtered = filtered.filter((poi) =>
        poi.tags?.some((tag) => poiTagFilterIds.includes(tag.id))
      );
    }
    return filtered;
  }, [pointsOfInterest, poiSearch, poiZoneFilterId, zonePoiMap, poiTagFilterIds]);

  const archetypeByPoiId = useMemo(() => {
    const result: Record<string, LocationArchetype> = {};
    if (!pointsOfInterest.length || !locationArchetypes.length) return result;
    pointsOfInterest.forEach((poi) => {
      const match = matchLocationArchetypeForPoi(poi, locationArchetypes);
      if (match) {
        result[poi.id] = match;
      }
    });
    return result;
  }, [pointsOfInterest, locationArchetypes]);

  useEffect(() => {
    if (!selectedQuest?.nodes?.length) return;
    if (!locationArchetypes.length) return;

    setChallengeDrafts((prev) => {
      let changed = false;
      const next = { ...prev };
      selectedQuest.nodes?.forEach((node) => {
        if (!node.pointOfInterestId) return;
        const match = archetypeByPoiId[node.pointOfInterestId];
        if (!match) return;
        const existing = next[node.id];
        if (existing?.locationArchetypeId) return;
        next[node.id] = {
          ...emptyChallengeForm,
          ...existing,
          locationArchetypeId: match.id,
        };
        changed = true;
      });
      return changed ? next : prev;
    });
  }, [archetypeByPoiId, locationArchetypes.length, selectedQuest?.nodes]);

  const questNodePoints = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return [];
    return selectedQuest.nodes
      .filter((node) => node.pointOfInterestId)
      .map((node) => {
        const poi = pointsOfInterest.find((item) => item.id === node.pointOfInterestId);
        if (!poi) return null;
        const lng = Number(poi.lng);
        const lat = Number(poi.lat);
        if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
        return {
          id: node.id,
          name: poi.name,
          orderIndex: node.orderIndex,
          lng,
          lat,
        };
      })
      .filter((entry): entry is { id: string; name: string; orderIndex: number; lng: number; lat: number } => Boolean(entry));
  }, [pointsOfInterest, selectedQuest?.nodes]);

  const questNodePoiIdSet = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return new Set<string>();
    return new Set(
      selectedQuest.nodes
        .map((node) => node.pointOfInterestId)
        .filter((id): id is string => Boolean(id))
    );
  }, [selectedQuest?.nodes]);

  const snapToZone = () => {
    if (!questMap.current || !questForm.zoneId) return;
    const zone =
      zoneDetailsById[questForm.zoneId] ||
      zones.find((z) => z.id === questForm.zoneId);
    if (!zone) return;
    const boundaryCoords =
      zone.boundaryCoords?.map((coord) => [coord.longitude, coord.latitude]) ??
      (Array.isArray(zone.boundary) ? zone.boundary : []);
    if (boundaryCoords.length > 0) {
      const [lng, lat] = boundaryCoords[0];
      questMap.current.setCenter([lng, lat]);
      questMap.current.setZoom(14);
      return;
    }
    const lng = typeof zone.longitude === 'number' ? zone.longitude : Number(zone.longitude);
    const lat = typeof zone.latitude === 'number' ? zone.latitude : Number(zone.latitude);
    if (!Number.isNaN(lng) && !Number.isNaN(lat)) {
      questMap.current.setCenter([lng, lat]);
      questMap.current.setZoom(12);
    }
  };

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;

    questNodeMarkers.current.forEach((marker) => marker.remove());
    poiMarkers.current.forEach((marker) => marker.remove());
    characterLocationMarkers.current.forEach((marker) => marker.remove());
    questNodeMarkers.current = [];
    poiMarkers.current = [];
    characterLocationMarkers.current = [];

    const bounds = new mapboxgl.LngLatBounds();
    let hasBounds = false;

    questNodePoints.forEach((point) => {
      const el = document.createElement('div');
      el.style.width = '12px';
      el.style.height = '12px';
      el.style.borderRadius = '9999px';
      el.style.background = '#f59e0b';
      el.style.border = '2px solid #92400e';
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([point.lng, point.lat])
        .setPopup(new mapboxgl.Popup({ offset: 12 }).setText(`Node ${point.orderIndex}: ${point.name}`))
        .addTo(questMap.current!);
      questNodeMarkers.current.push(marker);
      bounds.extend([point.lng, point.lat]);
      hasBounds = true;
    });

    const characterLocations = selectedCharacterLocations || [];
    const poiNearCharacter = (poi: PointOfInterest) => {
      const lng = Number(poi.lng);
      const lat = Number(poi.lat);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return false;
      return characterLocations.some((loc) => {
        const dLng = Math.abs(lng - loc.longitude);
        const dLat = Math.abs(lat - loc.latitude);
        return dLng < 0.0001 && dLat < 0.0001;
      });
    };

    characterLocations.forEach((loc) => {
      if (Number.isNaN(loc.latitude) || Number.isNaN(loc.longitude)) return;
      const el = document.createElement('div');
      el.style.width = '14px';
      el.style.height = '14px';
      el.style.borderRadius = '9999px';
      el.style.background = '#10b981';
      el.style.border = '2px solid #065f46';
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([loc.longitude, loc.latitude])
        .addTo(questMap.current!);
      characterLocationMarkers.current.push(marker);
      bounds.extend([loc.longitude, loc.latitude]);
      hasBounds = true;
    });

    filteredPointsOfInterest.forEach((poi) => {
      if (questNodePoiIdSet.has(poi.id)) return;
      const lng = Number(poi.lng);
      const lat = Number(poi.lat);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return;
      const el = document.createElement('div');
      el.style.width = '10px';
      el.style.height = '10px';
      el.style.borderRadius = '9999px';
      const hasCharacter = poiNearCharacter(poi);
      el.style.background = hasCharacter ? '#a855f7' : '#3b82f6';
      el.style.border = hasCharacter ? '2px solid #6b21a8' : '2px solid #1e40af';
      el.style.cursor = 'pointer';
      el.addEventListener('click', () => {
        setSelectedPoiForModal(poi);
      });
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([lng, lat])
        .addTo(questMap.current!);
      poiMarkers.current.push(marker);
      bounds.extend([lng, lat]);
      hasBounds = true;
    });

    const polygonFeatures = questPolygons
      .map((polygon) => {
        if (!polygon.coordinates) return null;
        polygon.coordinates.forEach((ring) => {
          ring.forEach((coord) => {
            bounds.extend([coord[0], coord[1]]);
            hasBounds = true;
          });
        });
        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: polygon.coordinates,
          },
          properties: {
            id: polygon.id,
          },
        };
      })
      .filter(Boolean);

    const sourceId = 'quest-polygons';
    const fillLayerId = 'quest-polygons-fill';
    const lineLayerId = 'quest-polygons-line';
    const existingSource = questMap.current.getSource(sourceId);
    if (existingSource) {
      (existingSource as mapboxgl.GeoJSONSource).setData({
        type: 'FeatureCollection',
        features: polygonFeatures as GeoJSON.Feature[],
      });
    } else {
      questMap.current.addSource(sourceId, {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: polygonFeatures as GeoJSON.Feature[],
        },
      });
      questMap.current.addLayer({
        id: fillLayerId,
        type: 'fill',
        source: sourceId,
        paint: {
          'fill-color': '#f59e0b',
          'fill-opacity': 0.15,
        },
      });
      questMap.current.addLayer({
        id: lineLayerId,
        type: 'line',
        source: sourceId,
        paint: {
          'line-color': '#f59e0b',
          'line-width': 2,
        },
      });
    }

    const zone =
      questForm.zoneId
        ? zoneDetailsById[questForm.zoneId] || zones.find((z) => z.id === questForm.zoneId)
        : null;
    const zoneBoundaryCoords =
      zone?.boundaryCoords?.map((coord) => [coord.longitude, coord.latitude]) ??
      (Array.isArray(zone?.boundary) ? zone?.boundary : []);

    if (zoneBoundaryCoords.length > 0) {
      console.log('Quest Map: fitting to zone boundary', zoneBoundaryCoords);
      const map = questMap.current;
      map.resize();
      zoneBoundaryCoords.forEach((coord) => {
        bounds.extend([coord[0], coord[1]]);
      });
      const fit = () => map.fitBounds(bounds, { padding: 40, maxZoom: 15 });
      requestAnimationFrame(fit);
      setTimeout(fit, 100);
      return;
    }

    if (questNodePoints.length > 0) {
      const firstNode = [...questNodePoints].sort((a, b) => a.orderIndex - b.orderIndex)[0];
      questMap.current.setCenter([firstNode.lng, firstNode.lat]);
      questMap.current.setZoom(14);
      return;
    }

    if (hasBounds) {
      questMap.current.fitBounds(bounds, { padding: 40, maxZoom: 15 });
    } else if (questForm.zoneId && zone) {
      if (zone) {
        const lng = typeof zone.longitude === 'number' ? zone.longitude : Number(zone.longitude);
        const lat = typeof zone.latitude === 'number' ? zone.latitude : Number(zone.latitude);
        if (!Number.isNaN(lng) && !Number.isNaN(lat)) {
          questMap.current.setCenter([lng, lat]);
          questMap.current.setZoom(12);
        }
      }
    }
  }, [filteredPointsOfInterest, questForm.zoneId, questMapLoaded, questNodePoiIdSet, questNodePoints, questPolygons, selectedCharacterLocations, zones, zoneDetailsById]);

  const updateQuestState = (questId: string, updater: (quest: Quest) => Quest) => {
    setQuests((prev) => prev.map((quest) => (quest.id === questId ? updater(quest) : quest)));
  };

  const handleCreateQuest = async () => {
    try {
      const payload = {
        name: questForm.name,
        description: questForm.description,
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        gold: Number(questForm.gold) || 0,
        itemRewards: questForm.itemRewards
          .map((reward) => ({
            inventoryItemId: Number(reward.inventoryItemId) || 0,
            quantity: Number(reward.quantity) || 0,
          }))
          .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0),
      };
      const created = await apiClient.post<Quest>('/sonar/quests', payload);
      setQuests((prev) => [created, ...prev]);
      setSelectedQuestId(created.id);
      setQuestForm({ ...emptyQuestForm });
      setShowCreateQuest(false);
    } catch (error) {
      console.error('Failed to create quest', error);
      alert('Failed to create quest. Please check the required fields.');
    }
  };

  const handleUpdateQuest = async () => {
    if (!selectedQuest) return;
    try {
      const payload = {
        name: questForm.name,
        description: questForm.description,
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        gold: Number(questForm.gold) || 0,
        itemRewards: questForm.itemRewards
          .map((reward) => ({
            inventoryItemId: Number(reward.inventoryItemId) || 0,
            quantity: Number(reward.quantity) || 0,
          }))
          .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0),
      };
      const updated = await apiClient.patch<Quest>(`/sonar/quests/${selectedQuest.id}`, payload);
      updateQuestState(selectedQuest.id, () => updated);
    } catch (error) {
      console.error('Failed to update quest', error);
      alert('Failed to update quest.');
    }
  };

  const handleSelectQuest = (quest: Quest) => {
    setSelectedQuestId(quest.id);
    setQuestForm({
      name: quest.name ?? '',
      description: quest.description ?? '',
      imageUrl: quest.imageUrl ?? '',
      zoneId: quest.zoneId ?? '',
      questGiverCharacterId: quest.questGiverCharacterId ?? '',
      questArchetypeId: quest.questArchetypeId ?? '',
      gold: quest.gold ?? 0,
      itemRewards: (quest.itemRewards ?? []).map((reward) => ({
        inventoryItemId: reward.inventoryItemId ? String(reward.inventoryItemId) : '',
        quantity: reward.quantity ?? 1,
      })),
    });
  };

  const handleAddQuestReward = () => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { ...emptyQuestReward }],
    }));
  };

  const handleUpdateQuestReward = (index: number, updates: Partial<{ inventoryItemId: string; quantity: number }>) => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.map((reward, rewardIndex) =>
        rewardIndex === index ? { ...reward, ...updates } : reward
      ),
    }));
  };

  const handleRemoveQuestReward = (index: number) => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.filter((_, rewardIndex) => rewardIndex !== index),
    }));
  };

  const handleCreateNode = async () => {
    if (!selectedQuest) return;
    try {
      const polygonPoints = nodeForm.nodeType === 'polygon' ? parsePolygonPoints(nodeForm.polygonPoints) : null;
      if (nodeForm.nodeType === 'polygon' && !polygonPoints) {
        alert('Please enter polygon points as JSON: [[lng,lat],[lng,lat],...]');
        return;
      }
      const payload = {
        orderIndex: Number(nodeForm.orderIndex) || 1,
        pointOfInterestId: nodeForm.nodeType === 'poi' ? nodeForm.pointOfInterestId || null : null,
        polygonPoints: nodeForm.nodeType === 'polygon' ? polygonPoints : undefined,
      };
      const created = await apiClient.post<QuestNode>('/sonar/questNodes', {
        ...payload,
        questId: selectedQuest.id,
      });
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        nodes: [...(quest.nodes ?? []), created].sort((a, b) => a.orderIndex - b.orderIndex),
      }));
      setNodeForm({ ...emptyNodeForm, orderIndex: (selectedQuest.nodes?.length ?? 0) + 2 });
    } catch (error) {
      console.error('Failed to create quest node', error);
      alert('Failed to create quest node.');
    }
  };

  const handleChallengeDraftChange = (nodeId: string, updates: Partial<typeof emptyChallengeForm>) => {
    setChallengeDrafts((prev) => ({
      ...prev,
      [nodeId]: { ...emptyChallengeForm, ...prev[nodeId], ...updates },
    }));
  };

  const handleCreateChallenge = async (node: QuestNode) => {
    const draft = challengeDrafts[node.id] ?? emptyChallengeForm;
    if (!draft.question.trim()) {
      alert('Please enter a question for this challenge.');
      return;
    }
    try {
      const payload = {
        tier: Number(draft.tier) || 1,
        question: draft.question,
        reward: Number(draft.reward) || 0,
        inventoryItemId: draft.inventoryItemId ? Number(draft.inventoryItemId) : null,
      };
      const created = await apiClient.post<QuestNodeChallenge>(`/sonar/questNodes/${node.id}/challenges`, payload);
      updateQuestState(node.questId, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).map((n) =>
          n.id === node.id ? { ...n, challenges: [...(n.challenges ?? []), created] } : n
        ),
      }));
      setChallengeDrafts((prev) => ({ ...prev, [node.id]: { ...emptyChallengeForm } }));
    } catch (error) {
      console.error('Failed to create challenge', error);
      alert('Failed to create challenge.');
    }
  };

  const handleDeleteNode = async (node: QuestNode) => {
    if (!selectedQuest) return;
    const confirmDelete = window.confirm(`Delete quest node ${node.orderIndex}? This cannot be undone.`);
    if (!confirmDelete) return;
    try {
      await apiClient.delete(`/sonar/questNodes/${node.id}`);
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).filter((n) => n.id !== node.id),
      }));
    } catch (error) {
      console.error('Failed to delete quest node', error);
      alert('Failed to delete quest node.');
    }
  };

  const resetImportForm = () => {
    setImportQuery('');
    setSelectedCandidate(null);
    setImportError(null);
    setImportJobs([]);
    setImportPolling(false);
    setImportZoneId(questForm.zoneId || '');
  };

  const handleImportPointOfInterest = async () => {
    setImportError(null);
    if (!selectedCandidate) {
      setImportError('Please select a Google Maps location.');
      return;
    }
    const zoneId = importZoneId || questForm.zoneId;
    if (!zoneId) {
      setImportError('Please select a zone.');
      return;
    }
    try {
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: selectedCandidate.place_id,
        zoneID: zoneId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Error importing point of interest:', error);
      setImportError('Failed to import point of interest.');
    }
  };

  const handleRetryImport = async (placeId: string, zoneId: string) => {
    try {
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: placeId,
        zoneID: zoneId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Failed to retry import', error);
      setImportError('Failed to retry import.');
    }
  };

  const fetchImportJobs = async (zoneId?: string) => {
    try {
      const url = zoneId ? `/sonar/pointOfInterest/imports?zoneId=${zoneId}` : '/sonar/pointOfInterest/imports';
      const response = await apiClient.get<PointOfInterestImport[]>(url);
      setImportJobs(response);
      const hasPending = response.some((item) => item.status === 'queued' || item.status === 'in_progress');
      setImportPolling(hasPending);
    } catch (error) {
      console.error('Failed to fetch import status', error);
    }
  };

  useEffect(() => {
    if (!showImportModal) return;
    fetchImportJobs(importZoneId || questForm.zoneId || undefined);
  }, [showImportModal, importZoneId, questForm.zoneId]);

  useEffect(() => {
    if (!importPolling) return;
    const interval = setInterval(() => {
      fetchImportJobs(importZoneId || questForm.zoneId || undefined);
    }, 3000);
    return () => clearInterval(interval);
  }, [importPolling, importZoneId, questForm.zoneId]);

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
          setNodeForm((form) => ({
            ...form,
            nodeType: 'poi',
            pointOfInterestId: job.pointOfInterestId || form.pointOfInterestId,
          }));
        }
      });
      return hasNew ? next : prev;
    });

    refreshPointsOfInterest();
  }, [importJobs]);

  const openCharacterLocations = async () => {
    if (!questForm.questGiverCharacterId) return;
    setCharacterLocationsOpen(true);
    setCharacterLocationsLoading(true);
    try {
      const response = await apiClient.get<{ latitude: number; longitude: number }[]>(
        `/sonar/characters/${questForm.questGiverCharacterId}/locations`
      );
      setSelectedCharacterLocations(response);
    } catch (error) {
      console.error('Failed to load character locations', error);
    } finally {
      setCharacterLocationsLoading(false);
    }
  };

  const handleAddCharacterLocation = () => {
    setSelectedCharacterLocations((prev) => [...prev, { latitude: 0, longitude: 0 }]);
  };

  const handleUpdateCharacterLocation = (index: number, key: 'latitude' | 'longitude', value: number) => {
    setSelectedCharacterLocations((prev) =>
      prev.map((loc, i) => (i === index ? { ...loc, [key]: value } : loc))
    );
  };

  const handleRemoveCharacterLocation = (index: number) => {
    setSelectedCharacterLocations((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSaveCharacterLocations = async () => {
    if (!questForm.questGiverCharacterId) return;
    try {
      await apiClient.put(`/sonar/characters/${questForm.questGiverCharacterId}/locations`, {
        locations: selectedCharacterLocations,
      });
      setCharacterLocationsOpen(false);
    } catch (error) {
      console.error('Failed to save character locations', error);
      alert('Failed to save character locations.');
    }
  };

  if (loading) {
    return <div className="m-10">Loading quests...</div>;
  }

  return (
    <div className="m-10">
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
      {loadError && (
        <div className="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {loadError}
        </div>
      )}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Quests</h1>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={() => setShowCreateQuest((prev) => !prev)}
        >
          {showCreateQuest ? 'Close' : 'Create Quest'}
        </button>
      </div>

      {showCreateQuest && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="text-lg font-semibold mb-3">Create Quest</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Name</label>
              <input
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.name}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700">Description</label>
              <textarea
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                rows={3}
                value={questForm.description}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, description: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Zone</label>
              <input
                className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                placeholder="Filter zones..."
                value={zoneSearch}
                onChange={(e) => setZoneSearch(e.target.value)}
              />
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.zoneId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, zoneId: e.target.value }))}
              >
                <option value="">No Zone</option>
                {filteredZones.map((zone) => (
                  <option key={zone.id} value={zone.id}>
                    {zone.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Quest Giver Character</label>
              <input
                className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                placeholder="Filter characters..."
                value={characterSearch}
                onChange={(e) => setCharacterSearch(e.target.value)}
              />
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.questGiverCharacterId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, questGiverCharacterId: e.target.value }))}
              >
                <option value="">None</option>
                {filteredCharacters.map((character) => (
                  <option key={character.id} value={character.id}>
                    {character.name}
                  </option>
                ))}
              </select>
              {questForm.questGiverCharacterId && (
                <button
                  type="button"
                  className="mt-2 rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                  onClick={openCharacterLocations}
                >
                  Edit Character Locations
                </button>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Quest Archetype ID (optional)</label>
              <input
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.questArchetypeId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, questArchetypeId: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Gold Reward</label>
              <input
                type="number"
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.gold}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, gold: Number(e.target.value) }))}
              />
            </div>
            <div className="md:col-span-2">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-gray-700">Item Rewards</label>
                <button
                  type="button"
                  className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                  onClick={handleAddQuestReward}
                >
                  Add Item Reward
                </button>
              </div>
              {questForm.itemRewards.length === 0 ? (
                <div className="mt-2 text-xs text-gray-500">No item rewards yet.</div>
              ) : (
                <div className="mt-2 space-y-2">
                  {questForm.itemRewards.map((reward, index) => (
                    <div
                      key={`create-reward-${index}`}
                      className="grid grid-cols-[1fr_120px_auto] gap-2 items-center"
                    >
                      <select
                        className="block w-full border border-gray-300 rounded-md p-2"
                        value={reward.inventoryItemId}
                        onChange={(e) => handleUpdateQuestReward(index, { inventoryItemId: e.target.value })}
                      >
                        <option value="">Select item</option>
                        {inventoryItems.map((item) => (
                          <option key={item.id} value={item.id}>
                            {item.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        className="block w-full border border-gray-300 rounded-md p-2"
                        min={1}
                        value={reward.quantity}
                        onChange={(e) => handleUpdateQuestReward(index, { quantity: Number(e.target.value) })}
                      />
                      <button
                        type="button"
                        className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                        onClick={() => handleRemoveQuestReward(index)}
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
          <div className="mt-4">
            <button
              className="bg-green-600 text-white px-4 py-2 rounded-md"
              onClick={handleCreateQuest}
              disabled={!questForm.name.trim()}
            >
              Create Quest
            </button>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-[340px_1fr] gap-6">
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-3">Quest List</h2>
          <input
            className="mb-3 block w-full border border-gray-300 rounded-md p-2"
            placeholder="Search quests..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <div className="space-y-2 max-h-[520px] overflow-y-auto">
            {filteredQuests.map((quest) => (
              <button
                key={quest.id}
                className={`w-full text-left p-3 rounded-md border ${selectedQuestId === quest.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}`}
                onClick={() => handleSelectQuest(quest)}
              >
                <div className="font-semibold">{quest.name}</div>
                <div className="text-xs text-gray-500">Nodes: {quest.nodes?.length ?? 0}</div>
              </button>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          {!selectedQuest ? (
            <div className="text-gray-500">Select a quest to edit details and add nodes.</div>
          ) : (
            <>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">Quest Details</h2>
                <div className="flex items-center gap-2">
                  <button
                    className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    onClick={() => {
                      resetImportForm();
                      setShowImportModal(true);
                    }}
                  >
                    Import POI
                  </button>
                  <button
                    className="bg-blue-600 text-white px-4 py-2 rounded-md"
                    onClick={handleUpdateQuest}
                  >
                    Save Changes
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Name</label>
                  <input
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.name}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, name: e.target.value }))}
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <textarea
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    rows={3}
                    value={questForm.description}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, description: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Zone</label>
                  <select
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.zoneId}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, zoneId: e.target.value }))}
                  >
                    <option value="">No Zone</option>
                    {zones.map((zone) => (
                      <option key={zone.id} value={zone.id}>
                        {zone.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Quest Giver Character</label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.questGiverCharacterId}
                  onChange={(e) => setQuestForm((prev) => ({ ...prev, questGiverCharacterId: e.target.value }))}
                >
                  <option value="">None</option>
                  {characters.map((character) => (
                    <option key={character.id} value={character.id}>
                      {character.name}
                    </option>
                  ))}
                </select>
                {questForm.questGiverCharacterId && (
                  <button
                    type="button"
                    className="mt-2 rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={openCharacterLocations}
                  >
                    Edit Character Locations
                  </button>
                )}
              </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Quest Archetype ID</label>
                  <input
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.questArchetypeId}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, questArchetypeId: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Gold Reward</label>
                  <input
                    type="number"
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.gold}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, gold: Number(e.target.value) }))}
                  />
                </div>
                <div className="md:col-span-2">
                  <div className="flex items-center justify-between">
                    <label className="block text-sm font-medium text-gray-700">Item Rewards</label>
                    <button
                      type="button"
                      className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                      onClick={handleAddQuestReward}
                    >
                      Add Item Reward
                    </button>
                  </div>
                  {questForm.itemRewards.length === 0 ? (
                    <div className="mt-2 text-xs text-gray-500">No item rewards yet.</div>
                  ) : (
                    <div className="mt-2 space-y-2">
                      {questForm.itemRewards.map((reward, index) => (
                        <div
                          key={`edit-reward-${index}`}
                          className="grid grid-cols-[1fr_120px_auto] gap-2 items-center"
                        >
                          <select
                            className="block w-full border border-gray-300 rounded-md p-2"
                            value={reward.inventoryItemId}
                            onChange={(e) => handleUpdateQuestReward(index, { inventoryItemId: e.target.value })}
                          >
                            <option value="">Select item</option>
                            {inventoryItems.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name}
                              </option>
                            ))}
                          </select>
                          <input
                            type="number"
                            className="block w-full border border-gray-300 rounded-md p-2"
                            min={1}
                            value={reward.quantity}
                            onChange={(e) => handleUpdateQuestReward(index, { quantity: Number(e.target.value) })}
                          />
                          <button
                            type="button"
                            className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                            onClick={() => handleRemoveQuestReward(index)}
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="text-lg font-semibold mb-3">Quest Nodes</h3>
                <div className="bg-gray-50 border border-gray-200 rounded-md p-4 mb-4">
                  <h4 className="font-semibold mb-3">Add Node</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Order Index</label>
                      <input
                        type="number"
                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                        value={nodeForm.orderIndex}
                        onChange={(e) => setNodeForm((prev) => ({ ...prev, orderIndex: Number(e.target.value) }))}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Node Type</label>
                      <select
                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                        value={nodeForm.nodeType}
                        onChange={(e) => setNodeForm((prev) => ({ ...prev, nodeType: e.target.value as 'poi' | 'polygon' }))}
                      >
                        <option value="poi">Point of Interest</option>
                        <option value="polygon">Polygon</option>
                      </select>
                    </div>
                    {nodeForm.nodeType === 'poi' ? (
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700">Point of Interest</label>
                        <input
                          className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                          placeholder="Search points of interest..."
                          value={poiSearch}
                          onChange={(e) => setPoiSearch(e.target.value)}
                        />
                        <button
                          type="button"
                          className="mb-2 flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
                          onClick={() => setPoiFiltersOpen((prev) => !prev)}
                        >
                          <span>Filters</span>
                          <span>{poiFiltersOpen ? 'Hide' : 'Show'}</span>
                        </button>
                        {poiFiltersOpen && (
                          <div className="mb-3 rounded-md border border-gray-200 bg-gray-50 p-3">
                            <div className="mb-3">
                              <label className="block text-xs font-medium text-gray-700">Zone</label>
                              <select
                                className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                value={poiZoneFilterId}
                                onChange={(e) => setPoiZoneFilterId(e.target.value)}
                              >
                                <option value="">All zones</option>
                                {zones.map((zone) => (
                                  <option key={zone.id} value={zone.id}>
                                    {zone.name}
                                  </option>
                                ))}
                              </select>
                              {poiZoneFilterId && zonePoiMapLoading && (
                                <p className="mt-1 text-xs text-gray-500">Loading zone points of interest…</p>
                              )}
                            </div>
                            <div>
                              <label className="block text-xs font-medium text-gray-700">Tags</label>
                              <input
                                className="mt-1 mb-2 block w-full rounded-md border border-gray-300 p-2"
                                placeholder="Search tags..."
                                value={poiTagSearch}
                                onChange={(e) => setPoiTagSearch(e.target.value)}
                              />
                              <div className="max-h-40 overflow-y-auto rounded-md border border-gray-200 bg-white p-2">
                                {filteredTags.length === 0 && (
                                  <div className="text-xs text-gray-500">No tags found.</div>
                                )}
                                {filteredTags.map((tag) => {
                                  const isSelected = poiTagFilterIds.includes(tag.id);
                                  return (
                                    <label key={tag.id} className="flex items-center gap-2 text-xs text-gray-700">
                                      <input
                                        type="checkbox"
                                        checked={isSelected}
                                        onChange={(e) => {
                                          if (e.target.checked) {
                                            setPoiTagFilterIds((prev) => [...prev, tag.id]);
                                          } else {
                                            setPoiTagFilterIds((prev) => prev.filter((id) => id !== tag.id));
                                          }
                                        }}
                                      />
                                      {tag.name}
                                    </label>
                                  );
                                })}
                              </div>
                            </div>
                            <div className="mt-3 flex items-center gap-2 text-xs">
                              <button
                                type="button"
                                className="rounded-md border border-gray-300 bg-white px-2 py-1"
                                onClick={() => {
                                  setPoiZoneFilterId('');
                                  setPoiTagFilterIds([]);
                                  setPoiTagSearch('');
                                }}
                              >
                                Clear filters
                              </button>
                              <span className="text-gray-500">
                                Showing {filteredPointsOfInterest.length} / {pointsOfInterest.length}
                              </span>
                            </div>
                          </div>
                        )}
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.pointOfInterestId}
                          onChange={(e) => setNodeForm((prev) => ({ ...prev, pointOfInterestId: e.target.value }))}
                        >
                          <option value="">Select a POI</option>
                          {filteredPointsOfInterest.map((poi) => (
                            <option key={poi.id} value={poi.id}>
                              {poi.name}
                            </option>
                          ))}
                        </select>
                      </div>
                    ) : (
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-gray-700">Polygon Points</label>
                        <textarea
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          rows={3}
                          placeholder='[[lng,lat],[lng,lat],[lng,lat]]'
                          value={nodeForm.polygonPoints}
                          onChange={(e) => {
                            const value = e.target.value;
                            setNodeForm((prev) => ({ ...prev, polygonPoints: value }));
                            const parsed = parsePolygonPoints(value);
                            if (parsed) {
                              setPolygonDraftPoints(parsed);
                            }
                          }}
                        />
                        <div className="mt-2 flex flex-wrap gap-2 text-xs text-gray-600">
                          <span className="rounded-md bg-teal-50 px-2 py-1 text-teal-700">
                            Click on the map to add polygon points.
                          </span>
                          <button
                            type="button"
                            className="rounded-md border border-gray-300 bg-white px-2 py-1 text-gray-700 hover:bg-gray-50"
                            onClick={() => {
                              setPolygonDraftPoints([]);
                              setNodeForm((prev) => ({ ...prev, polygonPoints: '' }));
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
                                setNodeForm((formPrev) => ({
                                  ...formPrev,
                                  polygonPoints: JSON.stringify(next),
                                }));
                                return next;
                              });
                            }}
                          >
                            Undo last point
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                  <button
                    className="mt-4 bg-green-600 text-white px-4 py-2 rounded-md"
                    onClick={handleCreateNode}
                    disabled={nodeForm.nodeType === 'poi' && !nodeForm.pointOfInterestId}
                  >
                    Add Node
                  </button>
                </div>

                <div className="mb-6 rounded-lg border border-gray-200 bg-white p-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-semibold">Quest Map</h4>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={snapToZone}
                      >
                        Snap to Zone
                      </button>
                      <button
                        type="button"
                        className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={() => setPolygonRefreshNonce((prev) => prev + 1)}
                      >
                        Add Polygons to Map
                      </button>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-600">
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-amber-500 border border-amber-800" />
                        Quest nodes
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-blue-500 border border-blue-800" />
                        Filtered POIs
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-purple-500 border border-purple-800" />
                        POIs with character
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500 border border-emerald-800" />
                        Character locations
                      </span>
                    </div>
                  </div>
                  <p className="mt-1 text-sm text-gray-600">
                    Use the POI filters above to narrow the map to locations that fit this quest.
                  </p>
                  <div className="mt-3 h-80 w-full overflow-hidden rounded-md border border-gray-200 relative">
                    <div ref={questMapContainer} className="h-full w-full" />
                  </div>
                </div>

                {selectedPoiForModal && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-lg">
                      <div className="flex items-start justify-between">
                        <h3 className="text-lg font-semibold">{selectedPoiForModal.name}</h3>
                        <button
                          className="text-gray-500 hover:text-gray-700"
                          onClick={() => setSelectedPoiForModal(null)}
                        >
                          Close
                        </button>
                      </div>
                      {selectedPoiForModal.imageURL && (
                        <img
                          src={selectedPoiForModal.imageURL}
                          alt={selectedPoiForModal.name}
                          className="mt-3 w-full rounded-md"
                        />
                      )}
                      {selectedPoiForModal.description && (
                        <p className="mt-3 text-sm text-gray-600">{selectedPoiForModal.description}</p>
                      )}
                      {selectedPoiForModal.tags && selectedPoiForModal.tags.length > 0 && (
                        <div className="mt-3 flex flex-wrap gap-2">
                          {selectedPoiForModal.tags.map((tag) => (
                            <span
                              key={tag.id}
                              className="rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-xs text-gray-700"
                            >
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      )}
                      <div className="mt-5 flex items-center justify-end gap-3">
                        <button
                          className="rounded-md border border-gray-300 px-4 py-2 text-sm"
                          onClick={() => setSelectedPoiForModal(null)}
                        >
                          Cancel
                        </button>
                        <button
                          className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
                          onClick={() => {
                            setNodeForm((prev) => ({
                              ...prev,
                              nodeType: 'poi',
                              pointOfInterestId: selectedPoiForModal.id,
                            }));
                            setSelectedPoiForModal(null);
                          }}
                        >
                          Select for Node
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {showImportModal && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-lg bg-white p-6 shadow-lg">
                      <div className="flex items-start justify-between">
                        <h3 className="text-lg font-semibold">Import Point of Interest</h3>
                        <button
                          className="text-gray-500 hover:text-gray-700"
                          onClick={() => {
                            setShowImportModal(false);
                            resetImportForm();
                          }}
                        >
                          Close
                        </button>
                      </div>

                      {importError && (
                        <div className="mt-3 text-sm text-red-600">{importError}</div>
                      )}

                      <div className="mt-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">Zone</label>
                        <select
                          className="w-full border border-gray-300 rounded-md px-3 py-2"
                          value={importZoneId || questForm.zoneId}
                          onChange={(e) => setImportZoneId(e.target.value)}
                        >
                          <option value="">Select a zone</option>
                          {zones.map((zone) => (
                            <option key={zone.id} value={zone.id}>
                              {zone.name}
                            </option>
                          ))}
                        </select>
                      </div>

                      <div className="mt-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">Search Google Maps</label>
                        <input
                          type="text"
                          className="w-full border border-gray-300 rounded-md px-3 py-2"
                          value={importQuery}
                          onChange={(e) => setImportQuery(e.target.value)}
                          placeholder="Search for a place..."
                        />
                      </div>

                      <div className="mt-4 border border-gray-200 rounded-md max-h-64 overflow-y-auto">
                        {candidates.length === 0 && (
                          <div className="p-4 text-sm text-gray-500">No results yet.</div>
                        )}
                        {candidates.map((candidate) => (
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
                          <h4 className="text-sm font-semibold">Import Status</h4>
                          <button
                            type="button"
                            className="text-xs text-blue-600"
                            onClick={() => fetchImportJobs(importZoneId || questForm.zoneId || undefined)}
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
                                    onClick={() => handleRetryImport(job.placeId, job.zoneId)}
                                  >
                                    Retry
                                  </button>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="mt-6 flex justify-end gap-2">
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

                {characterLocationsOpen && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-lg">
                      <div className="flex items-start justify-between">
                        <h3 className="text-lg font-semibold">Character Locations</h3>
                        <button
                          className="text-gray-500 hover:text-gray-700"
                          onClick={() => setCharacterLocationsOpen(false)}
                        >
                          Close
                        </button>
                      </div>
                      {characterLocationsLoading ? (
                        <div className="mt-4 text-sm text-gray-600">Loading locations...</div>
                      ) : (
                        <div className="mt-4 space-y-3">
                          {selectedCharacterLocations.length === 0 && (
                            <div className="text-sm text-gray-500">No locations yet.</div>
                          )}
                          {selectedCharacterLocations.map((location, index) => (
                            <div key={`${location.latitude}-${location.longitude}-${index}`} className="flex items-center gap-2">
                              <input
                                type="number"
                                className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                value={location.latitude}
                                onChange={(e) => handleUpdateCharacterLocation(index, 'latitude', Number(e.target.value))}
                                placeholder="Latitude"
                              />
                              <input
                                type="number"
                                className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                value={location.longitude}
                                onChange={(e) => handleUpdateCharacterLocation(index, 'longitude', Number(e.target.value))}
                                placeholder="Longitude"
                              />
                              <button
                                className="rounded-md border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-700"
                                onClick={() => handleRemoveCharacterLocation(index)}
                              >
                                Remove
                              </button>
                            </div>
                          ))}
                        </div>
                      )}
                      <div className="mt-4 flex items-center justify-between">
                        <button
                          className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                          onClick={handleAddCharacterLocation}
                        >
                          Add Location
                        </button>
                        <button
                          className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
                          onClick={handleSaveCharacterLocations}
                          disabled={characterLocationsLoading}
                        >
                          Save Locations
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                <div className="space-y-4">
                  {(selectedQuest.nodes ?? [])
                    .slice()
                    .sort((a, b) => a.orderIndex - b.orderIndex)
                    .map((node) => (
                      <div key={node.id} className="border border-gray-200 rounded-md p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-semibold">Node {node.orderIndex}</div>
                            <div className="text-sm text-gray-600">
                              {node.pointOfInterestId
                                ? `POI: ${pointsOfInterest.find((poi) => poi.id === node.pointOfInterestId)?.name ?? node.pointOfInterestId}`
                                : 'Polygon'}
                            </div>
                          </div>
                          <button
                            className="rounded-md border border-red-200 bg-red-50 px-3 py-1 text-xs text-red-700 hover:bg-red-100"
                            onClick={() => handleDeleteNode(node)}
                          >
                            Remove Node
                          </button>
                        </div>

                        <div className="mt-3">
                          <h4 className="font-semibold mb-2">Challenges</h4>
                          <div className="space-y-2 mb-3">
                            {(node.challenges ?? []).map((challenge) => (
                              <div key={challenge.id} className="border border-gray-200 rounded-md p-2 text-sm">
                                <div>Tier {challenge.tier} · Reward {challenge.reward}</div>
                                <div className="text-gray-600">{challenge.question}</div>
                              </div>
                            ))}
                            {(node.challenges ?? []).length === 0 && (
                              <div className="text-sm text-gray-500">No challenges yet.</div>
                            )}
                          </div>

                          <div className="bg-gray-50 border border-gray-200 rounded-md p-3">
                            <h5 className="font-semibold mb-2">Add Challenge</h5>
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                              {node.pointOfInterestId && (
                                <div className="md:col-span-4 rounded-md border border-amber-200 bg-amber-50 p-3">
                                  <div className="text-xs font-semibold text-amber-900">Location Archetype Challenge</div>
                                  <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-3">
                                    <div>
                                      <label className="block text-xs font-medium text-gray-700">Archetype</label>
                                      <select
                                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                        value={(challengeDrafts[node.id] ?? emptyChallengeForm).locationArchetypeId}
                                        onChange={(e) =>
                                          handleChallengeDraftChange(node.id, {
                                            locationArchetypeId: e.target.value,
                                            locationChallenge: '',
                                          })
                                        }
                                      >
                                        <option value="">Select archetype</option>
                                        {locationArchetypes.map((archetype) => (
                                          <option key={archetype.id} value={archetype.id}>
                                            {archetype.name}
                                          </option>
                                        ))}
                                      </select>
                                    </div>
                                    <div>
                                      <label className="block text-xs font-medium text-gray-700">Challenge</label>
                                      <select
                                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                        value={(challengeDrafts[node.id] ?? emptyChallengeForm).locationChallenge}
                                        onChange={(e) =>
                                          handleChallengeDraftChange(node.id, {
                                            locationChallenge: e.target.value,
                                            question: e.target.value,
                                          })
                                        }
                                      >
                                        <option value="">Select challenge</option>
                                        {locationArchetypes
                                          .find(
                                            (archetype) =>
                                              archetype.id ===
                                              (challengeDrafts[node.id] ?? emptyChallengeForm).locationArchetypeId
                                          )
                                          ?.challenges?.map((challenge) => (
                                            <option key={challenge} value={challenge}>
                                              {challenge}
                                            </option>
                                          ))}
                                      </select>
                                    </div>
                                  </div>
                                  <p className="mt-2 text-xs text-amber-800">
                                    Selecting a challenge will auto-fill the question field.
                                  </p>
                                </div>
                              )}
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Tier</label>
                                <input
                                  type="number"
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).tier}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { tier: Number(e.target.value) })}
                                />
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Reward</label>
                                <input
                                  type="number"
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).reward}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { reward: Number(e.target.value) })}
                                />
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Inventory Item</label>
                                <select
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).inventoryItemId}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { inventoryItemId: e.target.value })}
                                >
                                  <option value="">None</option>
                                  {inventoryItems.map((item) => (
                                    <option key={item.id} value={item.id}>
                                      {item.name}
                                    </option>
                                  ))}
                                </select>
                              </div>
                              <div className="md:col-span-4">
                                <label className="block text-xs font-medium text-gray-700">Question</label>
                                <textarea
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  rows={2}
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).question}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { question: e.target.value })}
                                />
                              </div>
                            </div>
                            <button
                              className="mt-3 bg-blue-600 text-white px-3 py-2 rounded-md"
                              onClick={() => handleCreateChallenge(node)}
                            >
                              Add Challenge
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
